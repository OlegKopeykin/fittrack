# FitTrack

Селф-хостед трекер тренировок: дневник, каталог упражнений с историей подходов,
программы (дни и предписания), прогресс веса тела. Один статический бинарь:
Go-бэкенд + встроенная React-SPA + SQLite.

## Стек

- **Backend**: Go, chi, sqlc, goose (миграции embedded), SQLite
  (`modernc.org/sqlite`, без CGO), cookie-сессии (scs) + персональные
  Bearer-токены, argon2id.
- **Frontend**: React 19 + TypeScript + Vite, TanStack Query, React Router,
  Tailwind CSS v4. Mobile-first; графики — самописный SVG-компонент, без внешних
  библиотек.
- **API**: `/api/v1` — JSON; cookie-сессии для веб-клиента и персональные
  Bearer-токены для интеграций (см. «Машинный API»).

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

CI собирает такой же бинарь на каждый push; на тег версии `vX.Y.Z` он
публикуется в GitHub Release — оттуда его можно скачать вместо локальной сборки.

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

### 5. Машинный API (опционально)

Помимо cookie-сессий веб-клиента, к `/api/v1` можно ходить с персональным
Bearer-токеном (заголовок `Authorization: Bearer <token>`). Токен выпускается
подкомандой (печатается один раз, в БД хранится только хэш):

```sh
sudo -u fittrack FITTRACK_DB=/var/lib/fittrack/fittrack.db \
  /opt/fittrack/bin/fittrack admin token --user <логин> --name <метка> [--days N]
```

`--days 0` (по умолчанию) — бессрочный. Отзыв — удалением строки из таблицы
`api_tokens`.

### 6. Автообновление (опционально)

В `deploy/` есть systemd-таймер, который периодически забирает свежий бинарь из
GitHub Release и обновляет сервис (проверка контрольной суммы, атомарная замена,
healthcheck, откат при неудаче). Установка — см. `deploy/README.md`.

## Лицензия

MIT
