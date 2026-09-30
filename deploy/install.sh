#!/usr/bin/env bash
# Установка нового бинаря FitTrack (push-деплой из CI). Запускается от root через
# sudo из deploy-entry.sh; бинарь приходит в stdin, аргументы — sha256 бинаря и
# git-sha коммита. Порядок: проверка sha256 → бэкап БД → атомарная замена →
# рестарт → healthcheck; при неудаче — откат на предыдущий бинарь.
set -euo pipefail

BIN="/opt/fittrack/bin/fittrack"
STATE="/var/lib/fittrack/deployed.sha"
HEALTH="http://127.0.0.1:8080/healthz"

want_sum="${1:-}"
commit="${2:-}"
[[ "$want_sum" =~ ^[0-9a-f]{64}$ ]] || { echo "usage: install.sh <sha256> <git-sha>" >&2; exit 2; }
[[ "$commit" =~ ^[0-9a-f]{40}$ ]]   || { echo "bad git sha" >&2; exit 2; }

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
cat > "$tmp"

got_sum="$(sha256sum "$tmp" | awk '{print $1}')"
[ "$got_sum" = "$want_sum" ] || { echo "sha256 не совпал: $got_sum" >&2; exit 1; }

healthy() {
    for _ in $(seq 1 15); do
        curl -fsS "$HEALTH" >/dev/null 2>&1 && return 0
        sleep 1
    done
    return 1
}

# Миграции БД накатываются при старте и назад не откатываются — снимок до замены.
systemctl start fittrack-backup.service

install -o root -g root -m 755 "$tmp" "${BIN}.new"
[ -f "$BIN" ] && cp -p "$BIN" "${BIN}.prev"
mv -f "${BIN}.new" "$BIN"
systemctl restart fittrack

if healthy; then
    echo "$commit" > "$STATE"
    echo "деплой $commit ок"
    exit 0
fi

echo "healthcheck не прошёл — откат на предыдущий бинарь" >&2
if [ -f "${BIN}.prev" ]; then
    mv -f "${BIN}.prev" "$BIN"
    systemctl restart fittrack
    healthy && echo "откат ок" >&2 || echo "после отката сервис тоже не отвечает" >&2
fi
exit 1
