#!/bin/sh
# Онлайн-снимок SQLite: консистентная копия без остановки сервиса.
# VACUUM INTO открывает read-транзакцию — при WAL не блокирует писателей.
# Хранит последние $KEEP снимков, старые удаляет. Если задан BACKUP_HOOK,
# вызывает его со снимком (например, для копии за пределы хоста).
set -eu

DB="${FITTRACK_DB:-/var/lib/fittrack/fittrack.db}"
DIR="${FITTRACK_BACKUP_DIR:-/var/lib/fittrack/backups}"
KEEP="${FITTRACK_BACKUP_KEEP:-14}"

mkdir -p "$DIR"
stamp="$(date -u +%Y%m%d-%H%M%SZ)"
snap="$DIR/fittrack-$stamp.db"

sqlite3 "$DB" "VACUUM INTO '$snap'"
gzip -f "$snap"
snap="$snap.gz"
echo "снимок: $snap ($(du -h "$snap" | cut -f1))"

# Необязательная выгрузка за пределы хоста: BACKUP_HOOK получает путь к снимку.
if [ -n "${FITTRACK_BACKUP_HOOK:-}" ]; then
	"$FITTRACK_BACKUP_HOOK" "$snap"
fi

# Ротация: оставляем последние $KEEP по имени (имя монотонно по времени).
count=$(ls -1 "$DIR"/fittrack-*.db.gz 2>/dev/null | wc -l | tr -d ' ')
if [ "$count" -gt "$KEEP" ]; then
	ls -1 "$DIR"/fittrack-*.db.gz | sort | head -n "$((count - KEEP))" | while IFS= read -r old; do
		rm -f "$old"
		echo "удалён старый: $old"
	done
fi
