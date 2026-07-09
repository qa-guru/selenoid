# Selenoid (qa-guru fork)

**Основной репозиторий экосистемы.** Форк [aerokube/selenoid](https://github.com/aerokube/selenoid) с нативной поддержкой Playwright.

[![Build Status](https://github.com/qa-guru/selenoid/workflows/build/badge.svg)](https://github.com/qa-guru/selenoid/actions?query=workflow%3Abuild)
[![Coverage](https://codecov.io/github/qa-guru/selenoid/coverage.svg)](https://codecov.io/gh/qa-guru/selenoid)
[![Go Report Card](https://goreportcard.com/badge/github.com/qa-guru/selenoid)](https://goreportcard.com/report/github.com/qa-guru/selenoid)
[![Release](https://img.shields.io/github/release/qa-guru/selenoid.svg)](https://github.com/qa-guru/selenoid/releases/latest)
[![Docker Pulls](https://img.shields.io/docker/pulls/qaguru/selenoid.svg)](https://hub.docker.com/r/qaguru/selenoid)
[![StackOverflow Tag](https://img.shields.io/badge/stackoverflow-selenoid-orange.svg?style=flat)](https://stackoverflow.com/questions/tagged/selenoid)

| | |
|---|---|
| **GitHub** | [qa-guru/selenoid](https://github.com/qa-guru/selenoid) |
| **Docker Hub** | [`qaguru/selenoid`](https://hub.docker.com/r/qaguru/selenoid) |

## Что это

Selenoid — лёгкий Selenium hub на Go, который поднимает браузеры в Docker-контейнерах. В этом форке добавлен **native Playwright**: клиент подключается по WebSocket, hub стартует Playwright-образ и проксирует соединение.

```
Тест (Selenium)   ──►  /wd/hub         ──►  qaguru/webdriver-chrome (Chrome)
Тест (Playwright) ──►  /playwright/... ──►  qaguru/playwright-* (browser-image)
```

## Экосистема

```
                    ┌─────────────┐
                    │     cm      │  установка на сервер
                    └──────┬──────┘
                           │ скачивает бинарники + browsers.json
                           ▼
┌──────────────┐    ┌─────────────────┐    ┌──────────────────────┐
│ selenoid-ui  │◄──►│ selenoid (hub)  │───►│ browser-image          │
│              │    │                 │    │ (browser containers) │
└──────────────┘    └────────┬────────┘    └──────────────────────┘
                             │
                             ▼
                    qaguru/webdriver-chrome (Selenium Chrome)
```

WebDriver Chrome — [`qaguru/webdriver-chrome`](https://hub.docker.com/r/qaguru/webdriver-chrome) (`browser-image/webdriver/`). Playwright-образы — [`qaguru/playwright-*`](https://hub.docker.com/u/qaguru). Hub — [`qaguru/selenoid`](https://hub.docker.com/r/qaguru/selenoid).

**Twilio** (`twilio/selenoid`) — legacy cold WebDriver-образы, **не Playwright**. Текущий стек: Microsoft MCR + qaguru wrapper для Playwright; Chrome for Testing + qaguru wrapper для WebDriver. Таблица: [docs/browser-versions.md](docs/browser-versions.md).

## Связанные репозитории

| GitHub | Роль |
|--------|------|
| **selenoid** (этот) | Hub — центральный компонент |
| [selenoid-ui](https://github.com/qa-guru/selenoid-ui) | UI для отладки; проксирует hub |
| [cm](https://github.com/qa-guru/cm) | Установщик: скачивает hub, UI, `browsers.json` |
| [browser-image](https://github.com/qa-guru/browser-image) | Docker-образы browser nodes (Playwright + WebDriver) |

## Native Playwright

WebSocket endpoint (схема URL совместима с Moon):

```
ws://localhost:4444/playwright/playwright-chromium/1.61.1?enableVNC=true&enableVideo=true
ws://127.0.0.1:4444/playwright/playwright-chromium/1.61.1?enableVNC=true&enableVideo=true
```

Документация и smoke-тест:

- [docs/playwright.md](docs/playwright.md)
- [examples/playwright](examples/playwright)
- [docs/browser-versions.md](docs/browser-versions.md) — политика стека, правила совместимости
- [docs/driver-versions-catalog.md](docs/driver-versions-catalog.md) — полный каталог версий (official / aerokube / Twilio / qaguru)

## Требования к окружению

| Компонент | Версия | Зачем |
|-----------|--------|-------|
| **Go** | **1.23.x** | Сборка hub (файл `.go-version` — `1.23.8`) |
| **Docker Engine** | **26.1.x** (рекомендуется 26.1.5) | Совпадает с `github.com/docker/docker v26.1.5` в hub |
| **Docker API** | **1.45** | Нативный протокол клиента Selenoid |

Проверка:

```bash
./scripts/check-toolchain.sh
docker version | grep -E 'Version:|API'
# Engine: 26.1.x, API version: 1.45
go version   # go1.23.x
```

На Mac Docker Desktop часто ставит Engine 27.x — hub работает с `DOCKER_API_VERSION=1.45` (скрипт `./scripts/start-selenoid.sh` выставляет автоматически). На Linux-сервере и в CI зафиксируйте Engine 26.1.x (`./scripts/docker-engine-pin-ubuntu.sh`). Engine 28+ и 29+ не рекомендуем — разрыв с Docker SDK в проекте.

Сборка без локального Go использует образ `golang:1.23` (см. `./scripts/build-selenoid.sh`).

## Сборка и запуск

```bash
./scripts/build-selenoid.sh
docker pull qaguru/webdriver-chrome:148 qaguru/webdriver-chrome:148-min qaguru/video-recorder:latest
docker pull qaguru/playwright-chromium:1.61.1   # или сборка в browser-image
./scripts/start-selenoid.sh
```

Или вручную:

```bash
go build -o selenoid .
DOCKER_API_VERSION=1.45 ./selenoid -conf config/browsers.json -limit 5
```

Конфигурация браузеров: канон — [`../dev/browsers.json`](../dev/browsers.json); в репо hub — [`config/browsers.json`](config/browsers.json) (синхронизация: `../dev/scripts/sync-cm-browsers.sh`).

Smoke-тест Playwright: [examples/playwright](examples/playwright) (`npm install && npm test`).

Подробности релизов hub: [docs/RELEASE_v2.1.0.md](docs/RELEASE_v2.1.0.md).

## browsers.json

Два стека — **WebDriver** (Selenium) и **Playwright** (WebSocket). В каждой группе три актуальные версии плюс legacy для курсов.

| Браузер в hub | Default | Версии | Docker-образ | Протокол |
|---------------|---------|--------|--------------|----------|
| `chrome` | `148.0` | 148.0, 148.0-min | `qaguru/webdriver-chrome:148`, `qaguru/webdriver-chrome:148-min` | WebDriver, `path: /` |
| `playwright-chromium` | `1.61.1` | 1.61.1, 1.61.1-min | `qaguru/playwright-chromium:<версия>` | Playwright |
| `playwright-firefox` | `1.61.1` | 1.61.1 | `qaguru/playwright-firefox:<версия>` | Playwright |
| `playwright-webkit` | `1.61.1` | 1.61.1 | `qaguru/playwright-webkit:<версия>` | Playwright |
| `playwright-chrome` | `1.61.1` | 1.61.1 | `qaguru/playwright-chrome:<версия>` | Playwright |
| `playwright-msedge` | `1.61.1` | 1.61.1 | `qaguru/playwright-msedge:<версия>` | Playwright |

Playwright-образы: порт `3000`, `protocol: "playwright"`, `shmSize: 2GB`. Версия npm-клиента `@playwright/test` должна совпадать с `playwrightVersion` в конфиге.

### Подготовка образов

```bash
# WebDriver Chrome (qaguru)
docker pull qaguru/webdriver-chrome:148 qaguru/webdriver-chrome:148-min

# Playwright — pull или сборка в browser-image
docker pull qaguru/playwright-chromium:1.61.1
# ./playwright/scripts/build.sh all 1.61.1   (в репозитории browser-image)

# Видеозапись сессий
docker pull qaguru/video-recorder:latest
```

### Endpoints (локально)

```
# WebDriver
http://127.0.0.1:4444/wd/hub

# Playwright
ws://127.0.0.1:4444/playwright/playwright-chromium/1.61.1?enableVNC=true&enableVideo=true
ws://127.0.0.1:4444/playwright/playwright-firefox/1.61.1?enableVNC=true&enableVideo=true
ws://127.0.0.1:4444/playwright/playwright-webkit/1.61.1?enableVNC=true&enableVideo=true
```

Записи с `enableVideo=true` сохраняются в каталог `video/` (или `http://127.0.0.1:4444/video/`).

## Переменные окружения

| Переменная | Описание |
|------------|----------|
| `DOCKER_API_VERSION` | API Docker для hub (рекомендуется `1.45`) |
| `PLAYWRIGHT_WS_ENDPOINT` | WebSocket URL hub |
| `PW_TEST_CONNECT_WS_ENDPOINT` | Alias (официальный env Playwright) |

## Upstream (Aerokube)

Форк [aerokube/selenoid](https://github.com/aerokube/selenoid) (upstream **UNMAINTAINED**). Оригинальная документация: [aerokube.com/selenoid/latest](http://aerokube.com/selenoid/latest/) и AsciiDoc в [`docs/*.adoc`](docs/) (upstream-only, не дублируют qa-guru quick start).
