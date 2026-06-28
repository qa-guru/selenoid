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
Тест (Selenium)   ──►  /wd/hub         ──►  twilio/selenoid (Chrome/Firefox/Edge)
Тест (Playwright) ──►  /playwright/... ──►  qaguru/playwright-* (см. playwright-image)
```

## Экосистема

```
                    ┌─────────────┐
                    │     cm      │  установка на сервер
                    └──────┬──────┘
                           │ скачивает бинарники + browsers.json
                           ▼
┌──────────────┐    ┌─────────────────┐    ┌──────────────────────┐
│ selenoid-ui  │◄──►│ selenoid (hub)  │───►│ playwright-image     │
│              │    │                 │    │ (browser containers) │
└──────────────┘    └────────┬────────┘    └──────────────────────┘
                             │
                             ▼
                    twilio/selenoid (Chrome/Firefox/Edge)
```

WebDriver-образы — [twilio/selenoid](https://hub.docker.com/r/twilio/selenoid). Playwright-образы — [`qaguru/playwright-*`](https://hub.docker.com/u/qaguru). Hub — [`qaguru/selenoid`](https://hub.docker.com/r/qaguru/selenoid).

## Связанные репозитории

| GitHub | Роль |
|--------|------|
| **selenoid** (этот) | Hub — центральный компонент |
| [selenoid-ui](https://github.com/qa-guru/selenoid-ui) | UI для отладки; проксирует hub |
| [cm](https://github.com/qa-guru/cm) | Установщик: скачивает hub, UI, `browsers.json` |
| [playwright-image](https://github.com/qa-guru/playwright-image) | Docker-образы browser nodes для Playwright |

## Native Playwright

WebSocket endpoint (схема URL совместима с Moon):

```
ws://localhost:4444/playwright/playwright-chromium/1.61.1?enableVNC=true&enableVideo=true
ws://127.0.0.1:4444/playwright/playwright-chromium/1.61.1?enableVNC=true&enableVideo=true
```

Документация и smoke-тест:

- [docs/playwright.md](docs/playwright.md)
- [examples/playwright](examples/playwright)
- [docs/browser-versions.md](docs/browser-versions.md) — таблица совместимости версий

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
docker pull twilio/selenoid:chrome_stable_148 twilio/selenoid:firefox_stable_150 selenoid/video-recorder:latest-release
docker pull qaguru/playwright-chromium:1.61.1   # или сборка в playwright-image
./scripts/start-selenoid.sh
```

Или вручную:

```bash
go build -o selenoid .
DOCKER_API_VERSION=1.45 ./selenoid -conf config/browsers.json -limit 5
```

Конфигурация браузеров: [`config/browsers.json`](config/browsers.json).

Smoke-тест Playwright: [examples/playwright](examples/playwright) (`npm install && npm test`).

Подробности релизов hub: [docs/RELEASE_v2.1.0.md](docs/RELEASE_v2.1.0.md).

## browsers.json

Два стека — **WebDriver** (Selenium) и **Playwright** (WebSocket). В каждой группе три актуальные версии плюс legacy для курсов.

| Браузер в hub | Default | Версии | Docker-образ | Протокол |
|---------------|---------|--------|--------------|----------|
| `chrome` | `148.0` | 148.0, 147.0, 146.0, 128.0 | `twilio/selenoid:chrome_stable_<N>`, `selenoid/vnc_chrome:128.0` | WebDriver, `path: /` |
| `firefox` | `150.0` | 150.0, 149.0, 148.0 | `twilio/selenoid:firefox_stable_<N>` | WebDriver, `path: /wd/hub` |
| `msedge` | `145.0` | 145.0, 144.0, 143.0 | `twilio/selenoid:edge_stable_<N>` | WebDriver, `path: /` |
| `playwright-chromium` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0, 1.46.0 | `qaguru/playwright-chromium:<версия>` | Playwright |
| `playwright-firefox` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0 | `qaguru/playwright-firefox:<версия>` | Playwright |
| `playwright-webkit` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0 | `qaguru/playwright-webkit:<версия>` | Playwright |
| `playwright-chrome` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0 | `qaguru/playwright-chrome:<версия>` | Playwright |
| `playwright-msedge` | `1.61.1` | 1.61.1, 1.61.0, 1.60.0 | `qaguru/playwright-msedge:<версия>` | Playwright |

Playwright-образы: порт `3000`, `protocol: "playwright"`, `shmSize: 2GB`. Версия npm-клиента `@playwright/test` должна совпадать с `playwrightVersion` в конфиге.

### Подготовка образов

```bash
# WebDriver (Twilio)
docker pull twilio/selenoid:chrome_stable_148 \
            twilio/selenoid:chrome_stable_147 \
            twilio/selenoid:chrome_stable_146 \
            twilio/selenoid:firefox_stable_150 \
            twilio/selenoid:firefox_stable_149 \
            twilio/selenoid:firefox_stable_148 \
            twilio/selenoid:edge_stable_145 \
            twilio/selenoid:edge_stable_144 \
            twilio/selenoid:edge_stable_143

# Playwright — pull или сборка в qa-guru/playwright-image
docker pull qaguru/playwright-chromium:1.61.1
# ./scripts/build.sh all 1.61.1   (в репозитории playwright-image)

# Видеозапись сессий
docker pull selenoid/video-recorder:latest-release
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

----------------------------------
## Оригинальная документация (Aerokube)

# Selenoid

[![Build Status](https://github.com/aerokube/selenoid/workflows/build/badge.svg)](https://github.com/aerokube/selenoid/actions?query=workflow%3Abuild)
[![Coverage](https://codecov.io/github/aerokube/selenoid/coverage.svg)](https://codecov.io/gh/aerokube/selenoid)
[![Go Report Card](https://goreportcard.com/badge/github.com/aerokube/selenoid)](https://goreportcard.com/report/github.com/aerokube/selenoid)
[![Release](https://img.shields.io/github/release/aerokube/selenoid.svg)](https://github.com/aerokube/selenoid/releases/latest)
[![Docker Pulls](https://img.shields.io/docker/pulls/aerokube/selenoid.svg)](https://hub.docker.com/r/aerokube/selenoid)
[![StackOverflow Tag](https://img.shields.io/badge/stackoverflow-selenoid-orange.svg?style=flat)](https://stackoverflow.com/questions/tagged/selenoid)

**UNMAINTAINED**. Consider https://aerokube.com/moon/latest as alternative.

Selenoid is a powerful implementation of [Selenium](http://github.com/SeleniumHQ/selenium) hub using [Docker](https://docker.com/) containers to launch browsers.
![Selenoid Animation](docs/img/selenoid-animation.gif)

## Features

### One-command Installation
Start browser automation in minutes by downloading [Configuration Manager](https://github.com/aerokube/cm/releases) binary and running just **one command**:
```
$ ./cm selenoid start --vnc --tmpfs 128
```
**That's it!** You can now use Selenoid instead of Selenium server. Specify the following Selenium URL in tests:
```
http://localhost:4444/wd/hub
```

### Ready to use Browser Images
No need to manually install browsers or dive into WebDriver documentation. Available images:
![Browsers List](docs/img/browsers-list.gif)

New images are added right after official releases. You can create your custom images with browsers. 

### Live Browser Screen and Logs
New **[rich user interface]((https://github.com/aerokube/selenoid-ui))** showing browser screen and Selenium session logs:
![Selenoid UI](docs/img/selenoid-ui.png)

### Video Recording
* Any browser session can be saved to [H.264](https://en.wikipedia.org/wiki/H.264/MPEG-4_AVC) video ([example](https://www.youtube.com/watch?v=maB298oO5cI))
* An API to list, download and delete recorded video files

### Convenient Logging

* Any browser session logs are automatically saved to files - one per session
* An API to list, download and delete saved log files

### Lightweight and Lightning Fast
Suitable for personal usage and in big clusters:
* Consumes **10 times** less memory than Java-based Selenium server under the same load
* **Small 6 Mb binary** with no external dependencies (no need to install Java)
* **Browser consumption API** working out of the box
* Ability to send browser logs to **centralized log storage** (e.g. to the [ELK-stack](https://logz.io/learn/complete-guide-elk-stack/))
* Fully **isolated** and **reproducible** environment

### Detailed Documentation and Free Support
Maintained by a growing community:
* Detailed [documentation](http://aerokube.com/selenoid/latest/)
* Telegram [support channel](https://t.me/aerokube)
* Support by [email](mailto:support@aerokube.com)
* StackOverflow [tag](https://stackoverflow.com/questions/tagged/selenoid)
* YouTube [channel](https://www.youtube.com/channel/UC9HvE3FNfTvftzpvXi9c69g)

## Complete Guide & Build Instructions

Complete reference guide (including building instructions) can be found at: http://aerokube.com/selenoid/latest/

## Selenoid in Kubernetes

Selenoid was initially created to be deployed on hardware servers or virtual machines and is not suitable for Kubernetes. Detailed motivation is described [here](https://aerokube.com/selenoid/latest/#_selenoid_in_kubernetes). If you still need running Selenium tests in Kubernetes, then take a look at [Moon](https://github.com/aerokube/moon/) - our dedicated solution for Kubernetes. 

## Known Users

[![JetBrains](docs/img/logo/jetbrains.png)](http://jetbrains.com/) [![Yandex](docs/img/logo/yandex.png)](https://yandex.com/company/) [![Sberbank Technology](docs/img/logo/sbertech.png)](http://sber-tech.com/) [![ThoughtWorks](docs/img/logo/thoughtworks.png)](https://thoughtworks.com/) [![VK.com](docs/img/logo/vk.png)](https://vk.com/) [![SuperJob](docs/img/logo/superjob.png)](http://superjob.ru/) [![PropellerAds](docs/img/logo/propellerads.png)](http://propellerads.com/) [![AlfaBank](docs/img/logo/alfabank.png)](https://alfabank.com/) [![3CX](docs/img/logo/3cx.png)](https://www.3cx.com/) [![IQ Option](docs/img/logo/iq_option.png)](https://iqoption.com/) [![Mail.Ru Group](docs/img/logo/mail_ru.png)](https://corp.mail.ru/en/) [![Newegg.Com](docs/img/logo/newegg.png)](https://newegg.com/) [![Badoo](docs/img/logo/badoo.png)](https://badoo.com/team/) [![BCS](docs/img/logo/bcs.png)](https://bcs.ru/) [![Quality Lab](docs/img/logo/quality-lab.png)](https://quality-lab.ru) [![AT Consulting](docs/img/logo/at-consulting.png)](https://www.at-consulting.ru/) [![Royal Caribbean International](docs/img/logo/royal-caribbean.png)](https://www.royalcaribbean.com/) [![Sixt](docs/img/logo/sixt.png)](https://sixt.com/) [![Testjar](docs/img/logo/testjar.png)](http://www.testjar.com/) [![Flipdish](docs/img/logo/flipdish.png)](https://www.flipdish.com/) [![RiAdvice](docs/img/logo/riadvice.png)](https://riadvice.tn/)
