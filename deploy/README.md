# Деплой

FitTrack — один статический бинарь (SPA встроена), SQLite лежит рядом.
Схема: Linux-хост, systemd-сервис на loopback, снаружи — реверс-прокси с TLS.
Обновление — **push-деплой из GitHub Actions** по SSH.

## Как обновляется прод

1. Push в рабочую ветку → `ci.yml`: Go/Web-тесты, сборка, E2E. `main` защищён
   (strict: ветка актуальна и прошла все проверки), поэтому после слияния тесты
   не повторяются.
2. Слияние в `main` → `deploy.yml`: сборка бинаря → SSH на хост под `deploy` →
   `install.sh`: проверка sha256, снимок БД (`fittrack-backup.service`),
   атомарная замена, рестарт, healthcheck; при неудаче — откат на прежний бинарь.
   Результат виден в Actions (environment `production`).
3. Если в `VERSION` новая версия (тега `vX.Y.Z` нет) — после деплоя
   публикуется GitHub Release. Теги руками ставить не нужно.

Ручной передеплой текущего `main`: Actions → Deploy → Run workflow.

Секреты репозитория: `DEPLOY_HOST` (адрес хоста), `DEPLOY_SSH_KEY` (приватный
ключ CI), `DEPLOY_KNOWN_HOSTS` (строка `ssh-keyscan` хоста).

## Подготовка хоста (однократно)

```sh
useradd --system --shell /usr/sbin/nologin fittrack
useradd -m -s /bin/bash deploy
mkdir -p /opt/fittrack/bin /var/lib/fittrack
chown fittrack:fittrack /var/lib/fittrack

install fittrack.service.example  /etc/systemd/system/fittrack.service
install -m 755 install.sh         /opt/fittrack/bin/install.sh
install -m 755 deploy-entry.sh    /opt/fittrack/bin/deploy-entry.sh
echo 'deploy ALL=(root) NOPASSWD: /opt/fittrack/bin/install.sh' > /etc/sudoers.d/deploy-fittrack

# Ключ CI умеет только передать бинарь (forced command, без shell и форвардинга):
echo 'command="/opt/fittrack/bin/deploy-entry.sh",restrict ssh-ed25519 AAAA… github-actions' \
  > ~deploy/.ssh/authorized_keys

systemctl daemon-reload
systemctl enable --now fittrack
```

Снимок перед деплоем требует установленного `fittrack-backup.service` (ниже).

## Первый owner-инвайт

```sh
sudo -u fittrack /opt/fittrack/bin/fittrack admin create-invite --owner
```

## Резервные копии и восстановление

Ежесуточный онлайн-снимок БД (`fittrack-backup.timer` → `backup.sh`):
консистентная копия через `VACUUM INTO`, gzip, хранятся последние 14 в
`/var/lib/fittrack/backups/`. Нужен `sqlite3` в системе.

```sh
install deploy/backup.sh                        /opt/fittrack/bin/backup.sh
install deploy/fittrack-backup.service.example  /etc/systemd/system/fittrack-backup.service
install deploy/fittrack-backup.timer.example    /etc/systemd/system/fittrack-backup.timer
systemctl daemon-reload && systemctl enable --now fittrack-backup.timer
```

Снять копию сейчас: `systemctl start fittrack-backup.service`.

**Восстановление (regularно проверять этот прогон):**

```sh
systemctl stop fittrack
cd /var/lib/fittrack
rm -f fittrack.db fittrack.db-wal fittrack.db-shm
gunzip -c backups/fittrack-<stamp>.db.gz > fittrack.db
chown fittrack:fittrack fittrack.db
systemctl start fittrack
curl -fsS http://127.0.0.1:8080/healthz
```

Снимки лежат на том же диске — это защита от повреждения БД, ошибочного
удаления и неудачной миграции, но **не** от потери диска или хоста. Для
устойчивости за пределами хоста задайте `FITTRACK_BACKUP_HOOK` (команда
получает путь к `.db.gz` — например, выгрузка в объектное хранилище) или
разверните потоковую репликацию (Litestream).

## Экспорт лога в Telegram (опционально)

Пользователи включают выгрузку своего лога в свой чат Telegram в профиле
(указывают токен бота). Ночной таймер обходит включивших и шлёт бэкап тем,
у кого подошёл срок по частоте (день/неделя/месяц). Нужен исходящий доступ
хоста к `api.telegram.org`.

```sh
install deploy/fittrack-telegram.service.example  /etc/systemd/system/fittrack-telegram.service
install deploy/fittrack-telegram.timer.example    /etc/systemd/system/fittrack-telegram.timer
systemctl daemon-reload && systemctl enable --now fittrack-telegram.timer
```

Прогнать вручную: `sudo -u fittrack FITTRACK_DB=/var/lib/fittrack/fittrack.db /opt/fittrack/bin/fittrack admin export-telegram`.
