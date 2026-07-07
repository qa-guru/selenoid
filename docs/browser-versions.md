# Таблица совместимости версий браузеров

Актуально для [`config/browsers.json`](../config/browsers.json). Последняя проверка: июль 2026.

## Политика стека: три слоя

| Слой | Upstream | Наш артефакт | Примечание |
|------|----------|--------------|------------|
| **Hub** | [aerokube/selenoid](https://github.com/aerokube/selenoid) | [qa-guru/selenoid](https://github.com/qa-guru/selenoid) → `qaguru/selenoid` | Форк hub с native Playwright. **Не Twilio.** |
| **Playwright nodes** | [mcr.microsoft.com/playwright](https://mcr.microsoft.com/en-us/product/playwright/about) + npm [`@playwright/test`](https://www.npmjs.com/package/@playwright/test) | `qaguru/playwright-*` ([browser-image/playwright](https://github.com/qa-guru/browser-image)) | Upstream npm/MCR без форка; кастомизируем только Docker node (VNC, `launchServer`, warm API). |
| **WebDriver nodes** | [Chrome for Testing](https://googlechromelabs.github.io/chrome-for-testing/) | `qaguru/webdriver-chrome*` ([browser-image/webdriver](https://github.com/qa-guru/browser-image)) | Selenium `/wd/hub`. Цель — полностью на qaguru-образах. |

### Зачем фигурировал Twilio

[twilio/selenoid](https://hub.docker.com/r/twilio/selenoid) — форк cold-образов aerokube/selenoid для **Selenium WebDriver** (Chrome/Firefox), когда официальные `selenoid/chrome` отставали по версиям. К **Playwright отношения не имеет**: Playwright всегда Microsoft base + qaguru wrapper.

**Текущая политика (июль 2026):** в `browsers.json`, CI, fixtures и сборке образов `twilio/*` **не используется**. WebDriver → `qaguru/webdriver-chrome`; Playwright → `qaguru/playwright-*`.

---

В Selenoid используются **два независимых стека**. Версии в `browsers.json` для них означают разное:

| Стек | Браузеры в `browsers.json` | Протокол | Что означает поле `version` |
|------|----------------------------|----------|------------------------------|
| **WebDriver** | `chrome` | Selenium (`/wd/hub`) | Мажорная версия браузера (`148.0` → Chrome 148.x) |
| **Playwright** | `playwright-chromium`, `playwright-firefox`, `playwright-webkit`, `playwright-chrome`, `playwright-msedge` | WebSocket (`/playwright/...`) | Версия npm-пакета `@playwright/test` (`1.61.1`) |

Образы — [`browser-image`](https://github.com/qa-guru/browser-image): `qaguru/webdriver-chrome`, `qaguru/playwright-*`.

> **Полный каталог** всех версий по источникам (Microsoft, aerokube, Twilio, qaguru): [driver-versions-catalog.md](driver-versions-catalog.md).

Firefox и Edge в Selenium-стеке **не публикуются** — используйте `playwright-firefox` / `playwright-msedge`.

---

## Сводная таблица (`browsers.json`)

| Имя в hub | Default | Версии в конфиге | Docker-образ | Протокол |
|-----------|---------|------------------|--------------|----------|
| `chrome` | `148.0` | 148.0, 148.0-min | `qaguru/webdriver-chrome:148`, `qaguru/webdriver-chrome:148-min` | WebDriver |
| `playwright-chromium` | `1.61.1` | 1.61.1, 1.61.1-min | `qaguru/playwright-chromium:<версия>` | Playwright |
| `playwright-firefox` | `1.61.1` | 1.61.1 | `qaguru/playwright-firefox:<версия>` | Playwright |
| `playwright-webkit` | `1.61.1` | 1.61.1 | `qaguru/playwright-webkit:<версия>` | Playwright |
| `playwright-chrome` | `1.61.1` | 1.61.1 | `qaguru/playwright-chrome:<версия>` | Playwright |
| `playwright-msedge` | `1.61.1` | 1.61.1 | `qaguru/playwright-msedge:<версия>` | Playwright |

---

## WebDriver: Chrome (`qaguru/webdriver-chrome`)

| Версия в hub | Docker-тег | Chrome в контейнере | `path` | Endpoint |
|--------------|------------|---------------------|--------|----------|
| **148.0** *(default)* | `qaguru/webdriver-chrome:148` | Chrome 148.x (warm: VNC + warm API) | `/` | `POST /wd/hub/session`, `browserName: chrome` |
| 148.0-min | `qaguru/webdriver-chrome:148-min` | CfT 148.0.7778.96 (headless CI) | `/` | то же, `version: 148.0-min` |

**Клиент:** любой Selenium WebDriver. Версия Selenium-клиента не привязана к версии браузера.

---

## Playwright: движки внутри каждой версии

Данные из официального [`browsers.json`](https://github.com/microsoft/playwright/blob/main/packages/playwright-core/browsers.json) Microsoft Playwright.

### Chromium Playwright (`playwright-chromium`)

| Playwright в hub | `playwrightVersion` | Chromium в контейнере | npm-клиент | WebSocket endpoint |
|------------------|---------------------|------------------------|------------|-------------------|
| **1.61.1** *(default)* | `1.61.1` | **149.0.7827.55** (rev 1223) | `@playwright/test@1.61.1` | `ws://host:4444/playwright/playwright-chromium/1.61.1?enableVNC=true&enableVideo=true` |
| 1.61.1-min | `1.61.1` | **149.0.7827.55** (headless CI) | `@playwright/test@1.61.1` | `ws://host:4444/playwright/playwright-chromium/1.61.1-min` |

### Firefox Playwright (`playwright-firefox`)

| Playwright в hub | `playwrightVersion` | Firefox в контейнере | npm-клиент | WebSocket endpoint |
|------------------|---------------------|----------------------|------------|-------------------|
| **1.61.1** *(default)* | `1.61.1` | **151.0** (rev 1522) | `@playwright/test@1.61.1` | `ws://host:4444/playwright/playwright-firefox/1.61.1?enableVNC=true&enableVideo=true` |

### WebKit Playwright (`playwright-webkit`)

| Playwright в hub | `playwrightVersion` | WebKit в контейнере | npm-клиент | WebSocket endpoint |
|------------------|---------------------|---------------------|------------|-------------------|
| **1.61.1** *(default)* | `1.61.1` | **26.5** (rev 2287) | `@playwright/test@1.61.1` | `ws://host:4444/playwright/playwright-webkit/1.61.1?enableVNC=true&enableVideo=true` |

---

## Правила совместимости клиента

| Что проверять | Правило | Последствие при нарушении |
|---------------|---------|---------------------------|
| Playwright npm ↔ hub | `@playwright/test` **должен совпадать** с `playwrightVersion` и версией в URL | Ошибки протокола, несовместимость CDP/BiDi |
| Playwright ↔ WebDriver | **Нельзя** подставить `chrome:148.0` вместо `playwright-chromium:1.61.1` | Сессия не создастся |
| Образ Docker | Образ из `browsers.json` должен быть **скачан** до старта сессии | `No such image` при создании сессии |

---

## Сравнение движков: Playwright vs WebDriver

| Playwright | Chromium/Firefox (Playwright) | WebDriver qaguru | Комментарий |
|------------|--------------------------------|------------------|-------------|
| 1.61.1 | Chromium **149** | Chrome **148.0** | Разные мажорные линии CfT / Playwright |
| 1.61.1 | Firefox **151** | — | Firefox только через Playwright |

---

## Матрица образов Docker

| Компонент | Команда pull / build |
|-----------|----------------------|
| WebDriver Chrome | `docker pull qaguru/webdriver-chrome:148 qaguru/webdriver-chrome:148-min` |
| Playwright 1.61.1 | [browser-image](https://github.com/qa-guru/browser-image) `./playwright/scripts/build.sh all 1.61.1` или `docker pull qaguru/playwright-chromium:1.61.1` |
| Видеозапись | `docker pull selenoid/video-recorder:latest-release` |

Базовый слой Playwright: `mcr.microsoft.com/playwright:v<версия>-noble` (Ubuntu Noble). Образы `qaguru/playwright-*` публикуются с тегом `<версия>` (например `1.61.1`).

Полный каталог версий (включая legacy 1.60.0): [driver-versions-catalog.md](driver-versions-catalog.md).

---

## Быстрый выбор стека

| Задача | Браузер в hub | Версия по умолчанию |
|--------|---------------|---------------------|
| Selenium-тесты Chrome | `chrome` | `148.0` |
| Selenium headless CI | `chrome` | `148.0-min` |
| `@playwright/test` в CI | `playwright-chromium` | `1.61.1` |
| Playwright + Firefox | `playwright-firefox` | `1.61.1` |
| Playwright + Edge | `playwright-msedge` | `1.61.1` |
| Playwright + Safari-движок | `playwright-webkit` | `1.61.1` |
