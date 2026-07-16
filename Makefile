GO      ?= go
NPM     ?= npm
BINARY  ?= bin/fittrack

.PHONY: deps test test-web build-web build build-e2e e2e vet clean

# Первый шаг после клона: JS-зависимости и браузеры Playwright.
deps:
	cd web && $(NPM) ci
	cd e2e && $(NPM) ci && npx playwright install

test:
	$(GO) vet ./...
	$(GO) test -race -cover ./...

test-web:
	cd web && $(NPM) test -- --run

build-web:
	cd web && $(NPM) run build

build: build-web
	$(GO) build -tags embedweb -o $(BINARY) ./cmd/fittrack

# Бинарь с e2e-фикстурами (тег e2eseed) — только для Playwright, не для поставки.
build-e2e: build-web
	$(GO) build -tags "embedweb e2eseed" -o $(BINARY) ./cmd/fittrack

e2e: build-e2e
	cd e2e && npx playwright test

clean:
	rm -rf bin web/dist
