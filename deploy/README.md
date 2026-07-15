# Деплой

FitTrack — один статический бинарь (SPA встроена), SQLite лежит рядом.
Схема: Linux-хост, systemd-сервис на loopback, снаружи — реверс-прокси с TLS.
Обновление — **pull-деплой**: хост сам забирает «катящийся» GitHub Release
(репозиторий публичный → без токенов и входящего SSH).

## Как обновляется прод

1. Зелёный `main` → job `release` в `ci.yml` публикует Release с тегом
   `rolling` (ассеты: `fittrack`, `fittrack.sha256`, `fittrack.sha`).
2. На хосте `fittrack-deploy.timer` раз в 2 минуты запускает
   `pull-deploy.sh`: сверяет `fittrack.sha` с установленным, при отличии —
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
