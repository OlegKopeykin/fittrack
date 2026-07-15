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

## Деплой

CI собирает статический linux/amd64-бинарь; деплой по тегу `v*` на Linux-хост
по SSH (systemd-сервис за реверс-прокси). Параметры — в GitHub Secrets,
шаблоны конфигов — в `deploy/`.

## Лицензия

MIT
