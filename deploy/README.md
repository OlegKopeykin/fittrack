# Деплой

FitTrack — один статический бинарь (SPA встроена), SQLite лежит рядом.
Схема: Linux-хост, systemd-сервис на loopback, снаружи — реверс-прокси с TLS.
Обновление — **pull-деплой**: хост сам забирает последний версионный GitHub
Release (репозиторий публичный → без токенов и входящего SSH).

## Как обновляется прод

1. Тег версии `vX.Y.Z` → job `release` в `ci.yml` собирает бинарь и публикует
   Release `vX.Y.Z` (ассеты: `fittrack`, `fittrack.sha256`, `fittrack.sha`).
   Push в `main` без тега только прогоняет тесты — прод не трогает.
2. На хосте `fittrack-deploy.timer` раз в 2 минуты запускает `pull-deploy.sh`:
   берёт последний релиз, сверяет `fittrack.sha` с установленным, при отличии —
   скачивает бинарь, проверяет sha256, атомарно заменяет, рестартит сервис,
   healthcheck. Sha фиксируется только после успешного healthcheck.

## Подготовка хоста (однократно)

```sh
useradd --system --shell /usr/sbin/nologin fittrack
mkdir -p /opt/fittrack/bin /var/lib/fittrack
chown fittrack:fittrack /var/lib/fittrack

install fittrack.service.example         /etc/systemd/system/fittrack.service
install pull-deploy.sh                    /opt/fittrack/bin/pull-deploy.sh
install fittrack-deploy.service.example   /etc/systemd/system/fittrack-deploy.service
install fittrack-deploy.timer.example     /etc/systemd/system/fittrack-deploy.timer

systemctl daemon-reload
systemctl enable --now fittrack
systemctl enable --now fittrack-deploy.timer
```

Реверс-прокси (TLS-терминация → `http://127.0.0.1:8080`) настраивается
отдельно под конкретный хост; в публичный репозиторий его конфиг не кладём.

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
