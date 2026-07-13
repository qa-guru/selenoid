# Selenoid (qa-guru fork)

<!-- stack-branches-note:start -->
> ## Стабильные билды — две ветки
>
> Стабильные версии стека зафиксированы в **двух долгоживущих ветках** (а не в `main`). Имя ветки кодирует согласованный toolchain всего стека, включая React из paired `selenoid-ui`:
>
> | Ветка | Стабильный билд | Docker API | Engine | Go | React | UI |
> |-------|-----------------|------------|--------|-----|-------|-----|
> | `selenoid2-1.45-engine26.1-go1.26-react16` | **v2.2.1** — прежний prod ([selenoid.autotests.cloud](https://selenoid.autotests.cloud)) | 1.45 | 26.1.x | 1.26.5 | 16 | CRA (react-scripts 3.x) |
> | `selenoid2-1.55-engine29.6-go1.26-react18` | **v2.3.0** — актуальный, до нового UI (Selenoid 3) | 1.55 | 29.6+ | 1.26.5 | 18 | Vite 6 |
>
> **Зачем две ветки:** каждая держит воспроизводимый набор версий (Docker API / Engine / Go / React). `main` — активная разработка. Точные версии — в `STACK-PIN.md`.
>
> _Вы на ветке `selenoid2-1.45-engine26.1-go1.26-react16`._
<!-- stack-branches-note:end -->


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
| **Текущий релиз** | **v2.2.1** — [docs/RELEASE_v2.2.1.md](docs/RELEASE_v2.2.1.md) · `qaguru/selenoid:v2.2.1` |

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

## Экосистема qa-guru Selenoid

| Ресурс | Ссылка | Роль |
|--------|--------|------|
| **selenoid** (этот) | [github.com/qa-guru/selenoid](https://github.com/qa-guru/selenoid) | Hub |
| selenoid-ui | [github.com/qa-guru/selenoid-ui](https://github.com/qa-guru/selenoid-ui) | Web UI |
| cm | [github.com/qa-guru/cm](https://github.com/qa-guru/cm) | Установщик |
| browser-image | [github.com/qa-guru/browser-image](https://github.com/qa-guru/browser-image) | Docker browser nodes |
| selenoid-tests | [github.com/qa-guru/selenoid-tests](https://github.com/qa-guru/selenoid-tests) | E2e/integration ethalon |
| Docker Hub | [hub.docker.com/u/qaguru](https://hub.docker.com/u/qaguru) | Образы `qaguru/*` |

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
| **Go** | **1.26.x** | Сборка hub (`.go-version` — `1.26.5`, `go.mod` `toolchain go1.26.5`) |
| **Docker Engine** | **26.1.x** (рекомендуется 26.1.5) | Совместим с `DOCKER_API_VERSION=1.45` (prod: Debian 10 / Engine 26.1.4) |
| **Docker API** | **1.45** | Версия API, с которой hub ходит к daemon (`DOCKER_API_VERSION`) |
| **Docker SDK** | **moby** | `github.com/moby/moby/client` + `github.com/moby/moby/api` (не `github.com/docker/docker`) |

Проверка:

```bash
./scripts/check-toolchain.sh
docker version | grep -E 'Version:|API'
# Engine: 26.1.x, API version: 1.45
go version   # go1.26.x
```

Hub и `cm` фиксируют `DOCKER_API_VERSION=1.45` (скрипт `./scripts/start-selenoid.sh`, образ, CI, prod). На Linux/CI: Engine **26.1.x** (`./scripts/docker-engine-pin-ubuntu.sh`). На Mac Docker Desktop (Engine 27.x) hub работает с тем же пином `1.45`.

Сборка без локального Go использует образ `golang:1.26` (см. `./scripts/build-selenoid.sh`).

## Сборка и запуск

```bash
./scripts/build-selenoid.sh
docker pull qaguru/webdriver-chrome:149 qaguru/webdriver-chrome:149-min qaguru/video-recorder:latest
docker pull qaguru/playwright-chromium:1.61.1   # или сборка в browser-image
./scripts/start-selenoid.sh
```

Или вручную:

```bash
go build -o selenoid .
DOCKER_API_VERSION=1.45 ./selenoid -conf config/browsers.json -limit 5
```

**SSOT:** [`../dev/browsers.json`](../dev/browsers.json) (полный каталог).  
Sync: [`../dev/scripts/sync-cm-browsers.sh`](../dev/scripts/sync-cm-browsers.sh) → [`config/browsers.json`](config/browsers.json), CM embed, CI fixture, UI.

Smoke-тест Playwright: [examples/playwright](examples/playwright) (`npm install && npm test`).

Текущий релиз hub: **v2.2.1** — [docs/RELEASE_v2.2.1.md](docs/RELEASE_v2.2.1.md) (patch: tag alignment + ecosystem README).  
Предыдущий: [docs/RELEASE_v2.2.0.md](docs/RELEASE_v2.2.0.md). Docker: `qaguru/selenoid:v2.2.1`.

## browsers.json

Два стека — **WebDriver** (Selenium) и **Playwright** (WebSocket). Каталог = **SSOT** [`../dev/browsers.json`](../dev/browsers.json). Полная таблица — [docs/browser-versions.md](docs/browser-versions.md).

| Браузер в hub | Default | Версии (sample) | Docker-образ | Протокол |
|---------------|---------|-----------------|--------------|----------|
| `chrome` | `149.0` | 149/148 (+`-min`) | `qaguru/webdriver-chrome:<tag>` | WebDriver, `path: /` |
| `firefox` | `151.0` | 151/150 (+`-min`) | `qaguru/webdriver-firefox:<tag>` | WebDriver |
| `msedge` | `145.0` | 145/144 (+`-min`) | `qaguru/webdriver-msedge:<tag>` | WebDriver |
| `playwright-chromium` | `1.61.1` | 1.61.1, 1.61.1-min | `qaguru/playwright-chromium:<версия>` | Playwright |
| `playwright-{firefox,webkit,chrome,msedge}` | `1.61.1` | 1.61.1 | `qaguru/playwright-*:<версия>` | Playwright |

Playwright-образы: порт `3000`, `protocol: "playwright"`, `shmSize: 2GB`. Версия npm-клиента `@playwright/test` должна совпадать с `playwrightVersion` в конфиге.

### Подготовка образов

```bash
# WebDriver Chrome (qaguru) — default 149.0; 148.* optional regression
docker pull qaguru/webdriver-chrome:149 qaguru/webdriver-chrome:149-min

# Playwright — pull или сборка в browser-image
docker pull qaguru/playwright-chromium:1.61.1
# ./playwright/scripts/build.sh all 1.61.1   (в репозитории browser-image)

# Видеозапись сессий
docker pull qaguru/video-recorder:latest
```

### Endpoints (локально, SSOT)

```
# WebDriver
http://127.0.0.1:4444/wd/hub

# Playwright
ws://127.0.0.1:4444/playwright/playwright-chromium/1.61.1?enableVNC=true&enableVideo=true
ws://127.0.0.1:4444/playwright/playwright-chromium/1.61.1-min
```

Записи с `enableVideo=true` сохраняются в каталог `video/` (или `http://127.0.0.1:4444/video/`).

## Переменные окружения (process hub)

| Переменная | Описание |
|------------|----------|
| `DOCKER_API_VERSION` | API Docker для hub (канон `1.45`) |
| `DOCKER_HOST` | Адрес Docker daemon (стандартный Docker client) |
| `GGR_HOST` | Хост GGR (если hub за GGR) |
| `OVERRIDE_VIDEO_OUTPUT_DIR` | Переопределение каталога video |

Клиентские `PLAYWRIGHT_WS_ENDPOINT` / `PW_TEST_CONNECT_WS_ENDPOINT` — не env процесса hub; см. [docs/playwright.md](docs/playwright.md) и [examples/playwright](examples/playwright).

## Upstream (Aerokube)

Форк [aerokube/selenoid](https://github.com/aerokube/selenoid) (upstream **UNMAINTAINED**). Оригинальная документация: [aerokube.com/selenoid/latest](http://aerokube.com/selenoid/latest/) и AsciiDoc в [`docs/*.adoc`](docs/) (upstream-only, не дублируют qa-guru quick start).
