#!/usr/bin/env bash
# Pull-деплой FitTrack: забирает «катящийся» GitHub Release и обновляет бинарь,
# только если sha изменился. Запускается по таймеру (fittrack-deploy.timer).
# Репозиторий публичный → GitHub API без токена; входящий SSH не нужен.
set -euo pipefail

REPO="${FITTRACK_REPO:-OlegKopeykin/fittrack}"
BIN="/opt/fittrack/bin/fittrack"
STATE="/var/lib/fittrack/deployed.sha"
API="https://api.github.com/repos/${REPO}/releases/tags/rolling"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# URL-ы ассетов свежего релиза (пока релиза нет — тихо выходим)
assets="$(curl -fsSL "$API" 2>/dev/null || true)"
if [ -z "$assets" ] || printf '%s' "$assets" | grep -q '"status": *"404"'; then
    echo "релиз rolling ещё не опубликован"
    exit 0
fi
# Строго browser_download_url, оканчивающийся точным именем ассета — иначе
# жадный grep поймает /fittrack из API-полей релиза (repos/.../fittrack).
url_for() {
    printf '%s\n' "$assets" \
        | grep -o '"browser_download_url": *"[^"]*"' \
        | sed -E 's/.*"(https[^"]*)"$/\1/' \
        | grep -E "/${1//./\\.}$" \
        | head -n1
}

sha_url="$(url_for fittrack.sha)"
[ -n "$sha_url" ] || { echo "нет ассета fittrack.sha"; exit 0; }
new_sha="$(curl -fsSL "$sha_url" | tr -d '[:space:]')"

if [ -f "$STATE" ] && [ "$(cat "$STATE")" = "$new_sha" ]; then
    exit 0   # уже актуально
fi

echo "новая сборка $new_sha — обновляюсь"
curl -fsSL "$(url_for fittrack)"        -o "$tmp/fittrack"
curl -fsSL "$(url_for fittrack.sha256)" -o "$tmp/fittrack.sha256"

# проверка целостности
echo "$(cat "$tmp/fittrack.sha256")  $tmp/fittrack" | sha256sum -c - >/dev/null

chmod +x "$tmp/fittrack"
install -o root -g root -m 755 "$tmp/fittrack" "${BIN}.new"
mv -f "${BIN}.new" "$BIN"          # атомарная замена
systemctl restart fittrack

# healthcheck; при неудаче — не фиксируем sha, таймер повторит
for _ in $(seq 1 15); do
    if curl -fsS http://127.0.0.1:8080/healthz >/dev/null; then
        echo "$new_sha" > "$STATE"
        echo "деплой $new_sha ок"
        exit 0
    fi
    sleep 1
done
echo "healthcheck не прошёл после обновления"
exit 1
