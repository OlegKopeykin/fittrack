# FitTrack

Селф-хостед трекер тренировок: дневник, каталог упражнений, программы с
прогрессией, графики прогресса. Один статический бинарь: Go-бэкенд + встроенная
React-SPA + SQLite.

## Стек

- **Backend**: Go, chi, sqlc, goose (миграции embedded), SQLite
  (`modernc.org/sqlite`, без CGO), cookie-сессии (scs), argon2id.
- **Frontend**: React 19 + TypeScript + Vite, TanStack Query, React Router,
  Tailwind CSS v4, Recharts. Mobile-first PWA.
- **API**: `/api/v1` — JSON; для интеграций — персональные Bearer-токены
  (контракт: `api/openapi.yaml`).

## Разработка

Проект ведётся по TDD: сначала тесты, потом код.

```sh
make deps       # первый шаг после клона: npm ci + браузеры Playwright
make test       # Go-тесты (-race -cover)
make test-web   # Vitest (web/)
make build      # web/dist + бинарь с встроенной SPA (тег embedweb)
make e2e        # Playwright против собранного бинаря (iPhone 15 Pro + Desktop Chrome)
```

Dev-режим: `go run ./cmd/fittrack` (API на `127.0.0.1:8080`) и `npm run dev`
в `web/` (Vite-прокси на API).

## Развёртывание на своём сервере

FitTrack — один статический бинарь: Go-бэкенд со встроенной React-SPA и файлом
SQLite рядом. Нужен Linux-сервер и реверс-прокси с TLS (nginx, Caddy и т.п.)
перед приложением. Шаблоны systemd-юнитов — в каталоге `deploy/`.

### 1. Собрать бинарь

Статический linux/amd64 со встроенной SPA:

```sh
make build   # web/dist + bin/fittrack
# или напрямую:
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags embedweb -o fittrack ./cmd/fittrack
```

CI на каждый push в `main` собирает такой же бинарь и публикует его в GitHub
Release — его можно скачать вместо локальной сборки.

### 2. Подготовить сервер

```sh
useradd --system --shell /usr/sbin/nologin fittrack
mkdir -p /opt/fittrack/bin /var/lib/fittrack
chown fittrack:fittrack /var/lib/fittrack

install fittrack                        /opt/fittrack/bin/fittrack
install deploy/fittrack.service.example /etc/systemd/system/fittrack.service
systemctl daemon-reload && systemctl enable --now fittrack
```

Приложение слушает loopback; настройки — через окружение в юните:

| Переменная | Назначение | По умолчанию |
|---|---|---|
| `FITTRACK_ADDR` | адрес прослушивания | `127.0.0.1:8080` |
| `FITTRACK_DB` | путь к файлу SQLite | `fittrack.db` |
| `FITTRACK_PUBLIC_ORIGIN` | внешний origin (`https://…`) — для Secure-кук и CSRF | — |

Миграции БД накатываются автоматически при старте.

### 3. Реверс-прокси и TLS

Направьте свой домен на сервер, терминируйте TLS реверс-прокси и проксируйте
запросы на `http://127.0.0.1:8080`. Пробрасывайте реальный IP клиента
(`X-Real-IP`) — на него завязан лимит попыток входа.

### 4. Первый вход

Регистрация закрыта — только по инвайтам. Создайте инвайт владельца:

```sh
sudo -u fittrack FITTRACK_DB=/var/lib/fittrack/fittrack.db \
  /opt/fittrack/bin/fittrack admin create-invite --owner
```

Откройте `https://ваш-домен/register?code=<код>` и заведите аккаунт. Сброс
пароля — `fittrack admin reset-password <логин>`.

### 5. Автообновление (опционально)

В `deploy/` есть systemd-таймер, который периодически забирает свежий бинарь из
GitHub Release и обновляет сервис (проверка контрольной суммы, атомарная замена,
healthcheck, откат при неудаче). Установка — см. `deploy/README.md`.

## Лицензия

MIT
