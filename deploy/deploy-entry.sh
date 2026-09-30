#!/usr/bin/env bash
# Forced command для ключа CI в ~deploy/.ssh/authorized_keys: ключ умеет только
# передать бинарь в install.sh. Клиент шлёт "<sha256> <git-sha>" как команду.
set -euo pipefail
read -r sum commit extra <<< "${SSH_ORIGINAL_COMMAND:-}"
[ -z "${extra:-}" ] || { echo "лишние аргументы" >&2; exit 2; }
exec sudo -n /opt/fittrack/bin/install.sh "${sum:-}" "${commit:-}"
