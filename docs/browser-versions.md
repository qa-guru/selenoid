# Таблица совместимости версий браузеров

**SSOT prod:** [qa-guru/browser-image](https://github.com/qa-guru/browser-image) `pins.json` + `watch.yml` → `catalog_sync.sh` (hub / cm / tests / UI + **последним** `deploy/browsers-production.json`). Box1: copy + `docker pull` + **SIGHUP** хабу — без stop hub/UI. Локально: [`../../dev/browsers.json`](../../dev/browsers.json). Сейчас: Chrome default **152.0**, Playwright **1.62.1**, Firefox **154.0**, Edge **151.0**. Последняя проверка: 26 августа 2026.

## Политика стека: три слоя

| Слой | Upstream | Наш артефакт | Примечание |
|------|----------|--------------|------------|
| **Hub** | [aerokube/selenoid](https://github.com/qa-guru/selenoid) | [qa-guru/selenoid](https://github.com/qa-guru/selenoid) → `qaguru/selenoid` | Форк hub с native Playwright. **Не Twilio.** |
| **Playwright nodes** | [mcr.microsoft.com/playwright](https://mcr.microsoft.com/en-us/product/playwright/about) + npm [`@playwright/test`](https://www.npmjs.com/package/@playwright/test) | `qaguru/playwright-*` ([browser-image/playwright](https://github.com/qa-guru/browser-image)) | Upstream npm/MCR без форка; кастомизируем только Docker node (VNC, `launchServer`, warm API). |
| **WebDriver nodes** | Chrome for Testing / Mozilla / Microsoft Edge | `qaguru/webdriver-chrome*` · `webdriver-firefox*` · `webdriver-msedge*` ([browser-image/webdriver](https://github.com/qa-guru/browser-image)) | Selenium `/wd/hub`, driver path `/`. Warm (VNC) + min (CI). |

### Зачем фигурировал Twilio

[twilio/selenoid](https://hub.docker.com/r/twilio/selenoid) — форк cold-образов aerokube/selenoid для **Selenium WebDriver** (Chrome/Firefox/Edge), когда официальные `selenoid/*` отставали по версиям. К **Playwright отношения не имеет**.

**Текущая политика (июль 2026):** в `browsers.json`, CI, fixtures и сборке образов `twilio/*` **не используется**. WebDriver → `qaguru/webdriver-{chrome,firefox,msedge}`; Playwright → `qaguru/playwright-*`.

---

В Selenoid используются **два независимых стека**. Версии в `browsers.json` для них означают разное:

| Стек | Браузеры в `browsers.json` | Протокол | Что означает поле `version` |
|------|----------------------------|----------|------------------------------|
| **WebDriver** | `chrome`, `firefox`, `msedge` | Selenium (`/wd/hub`) | Мажорная версия браузера (`152.0` → Chrome 152.x) |
| **Playwright** | `playwright-chromium`, `playwright-firefox`, `playwright-webkit`, `playwright-chrome`, `playwright-msedge` | WebSocket (`/playwright/...`) | Версия npm-пакета `@playwright/test` (`1.62.1`) |

Образы — [`browser-image`](https://github.com/qa-guru/browser-image).

> **Полный каталог** всех версий по источникам: [driver-versions-catalog.md](driver-versions-catalog.md).

---

## Сводная таблица (расширенный каталог)

| Имя в hub | Default | Версии в конфиге | Docker-образ | Протокол | В SSOT `dev/` |
|-----------|---------|------------------|--------------|----------|:-------------:|
| `chrome` | `152.0` | 152.0, 152.0-min, 151.0, 151.0-min, 148.0, 148.0-min | `qaguru/webdriver-chrome:<major>[-min]` | WebDriver | ✓ |
| `firefox` | `154.0` | 154.0, 154.0-min, 153.0, 153.0-min | `qaguru/webdriver-firefox:<major>[-min]` | WebDriver | ✓ |
| `msedge` | `151.0` | 151.0, 151.0-min, 150.0, 150.0-min | `qaguru/webdriver-msedge:<major>[-min]` | WebDriver | ✓ |
| `playwright-chromium` | `1.62.1` | 1.62.1, 1.62.1-min, 1.61.1, 1.61.1-min | `qaguru/playwright-chromium:<версия>` | Playwright | ✓ |
| `playwright-firefox` | `1.62.1` | 1.62.1, 1.61.1 | `qaguru/playwright-firefox:<версия>` | Playwright | ✓ |
| `playwright-webkit` | `1.62.1` | 1.62.1, 1.61.1 | `qaguru/playwright-webkit:<версия>` | Playwright | ✓ |
| `playwright-chrome` | `1.62.1` | 1.62.1, 1.61.1 | `qaguru/playwright-chrome:<версия>` | Playwright | ✓ |
| `playwright-msedge` | `1.62.1` | 1.62.1, 1.61.1 | `qaguru/playwright-msedge:<версия>` | Playwright | ✓ |

---

## WebDriver: Chrome / Firefox / Edge

| Версия в hub | Docker-тег | Содержимое | `path` | Endpoint |
|--------------|------------|------------|--------|----------|
| **chrome 152.0** *(default)* | `qaguru/webdriver-chrome:152` | CfT + chromedriver, warm VNC | `/` | `POST /wd/hub/session`, `browserName: chrome` |
| chrome 152.0-min | `qaguru/webdriver-chrome:152-min` | headless CI | `/` | то же, `browserVersion: 152.0-min` |
| chrome 151.0 / 151.0-min | `…:151` / `:151-min` | regression | `/` | то же |
| chrome 148.0 / 148.0-min | `…:148` / `:148-min` | compat (Jenkins/CFT pin) | `/` | то же |
| **firefox 154.0** *(default)* | `qaguru/webdriver-firefox:154` | Mozilla FF + geckodriver 0.37.1, warm VNC | `/` | `browserName: firefox` |
| firefox 154.0-min / 153.* | `…:154-min` / `:153[-min]` | CI / regression | `/` | то же |
| **msedge 151.0** *(default)* | `qaguru/webdriver-msedge:151` | Edge + msedgedriver, warm VNC (**amd64 only**) | `/` | `browserName: MicrosoftEdge` |
| msedge 151.0-min / 150.* | `…:151-min` / `:150[-min]` | CI / regression | `/` | то же |

**Клиент:** любой Selenium WebDriver. Edge в caps — `MicrosoftEdge` (alias → `msedge` в hub).

Каталожный суффикс `-min` (например `browserVersion: 151.0-min`) выбирает min-образ; hub перед прокси в driver переписывает версию в `151.0`, чтобы geckodriver / строгие драйверы не отвергали caps.

---

## Playwright: движки внутри каждой версии

Данные из официального [`browsers.json`](https://github.com/microsoft/playwright/blob/main/packages/playwright-core/browsers.json) Microsoft Playwright.

### Chromium Playwright (`playwright-chromium`)

| Playwright в hub | `playwrightVersion` | Chromium в контейнере | npm-клиент | WebSocket endpoint |
|------------------|---------------------|------------------------|------------|-------------------|
| **1.62.1** *(default)* | `1.62.1` | **151.0.7922.34** (rev 1234) | `@playwright/test@1.62.1` | `ws://host:4444/playwright/playwright-chromium/1.62.1?enableVNC=true&enableVideo=true` |
| 1.62.1-min | `1.62.1` | **151.0.7922.34** (headless CI) | `@playwright/test@1.62.1` | `ws://host:4444/playwright/playwright-chromium/1.62.1-min` |
| 1.61.1 / 1.61.1-min | `1.61.1` | **149.0.7827.55** (rev 1223) | `@playwright/test@1.61.1` | `…/playwright-chromium/1.61.1[-min]` |

### Firefox Playwright (`playwright-firefox`)

| Playwright в hub | `playwrightVersion` | Firefox в контейнере | npm-клиент | WebSocket endpoint |
|------------------|---------------------|----------------------|------------|-------------------|
| **1.62.1** *(default)* | `1.62.1` | **153.0** (rev 1538) | `@playwright/test@1.62.1` | `ws://host:4444/playwright/playwright-firefox/1.62.1?enableVNC=true&enableVideo=true` |
| 1.61.1 | `1.61.1` | **151.0** (rev 1522) | `@playwright/test@1.61.1` | `…/playwright-firefox/1.61.1` |

### WebKit Playwright (`playwright-webkit`)

| Playwright в hub | `playwrightVersion` | WebKit в контейнере | npm-клиент | WebSocket endpoint |
|------------------|---------------------|---------------------|------------|-------------------|
| **1.62.1** *(default)* | `1.62.1` | **26.5** (rev 2336) | `@playwright/test@1.62.1` | `ws://host:4444/playwright/playwright-webkit/1.62.1?enableVNC=true&enableVideo=true` |
| 1.61.1 | `1.61.1` | **26.5** (rev 2287) | `@playwright/test@1.61.1` | `…/playwright-webkit/1.61.1` |

---

## Правила совместимости клиента

| Что проверять | Правило | Последствие при нарушении |
|---------------|---------|---------------------------|
| Playwright npm ↔ hub | `@playwright/test` **должен совпадать** с `playwrightVersion` и версией в URL | Ошибки протокола, несовместимость CDP/BiDi |
| Playwright ↔ WebDriver | **Нельзя** подставить `chrome:152.0` вместо `playwright-chromium:1.62.1` | Сессия не создастся |
| Образ Docker | Образ из `browsers.json` должен быть **скачан** до старта сессии | `No such image` при создании сессии |

---

## Сравнение движков: Playwright vs WebDriver

| Playwright | Chromium/Firefox (Playwright) | WebDriver qaguru | Комментарий |
|------------|--------------------------------|------------------|-------------|
| 1.62.1 | Chromium **151** | Chrome **152.0** | Близкие CfT-линии; WD на milestone впереди |
| 1.62.1 | Firefox **153** | Firefox **154.0** | WebDriver и Playwright параллельны |

---

## Матрица образов Docker

| Компонент | Команда pull / build |
|-----------|----------------------|
| WebDriver Chrome | `docker pull qaguru/webdriver-chrome:152 qaguru/webdriver-chrome:152-min` |
| WebDriver Firefox | `docker pull qaguru/webdriver-firefox:154 qaguru/webdriver-firefox:154-min` |
| WebDriver Edge | `docker pull qaguru/webdriver-msedge:151 qaguru/webdriver-msedge:151-min` |
| Playwright 1.62.1 | [browser-image](https://github.com/qa-guru/browser-image) `./playwright/scripts/build.sh all 1.62.1` |
| Видеозапись | `docker pull qaguru/video-recorder:latest` · [browser-image/video-recorder](https://github.com/qa-guru/browser-image/tree/master/video-recorder) |

Базовый слой Playwright: `mcr.microsoft.com/playwright:v<версия>-noble` (Ubuntu Noble).

Полный каталог версий (включая legacy 1.60.0 на prod): [driver-versions-catalog.md](driver-versions-catalog.md).

---

## Быстрый выбор стека

| Задача | Браузер в hub | Версия по умолчанию |
|--------|---------------|---------------------|
| Selenium Chrome | `chrome` | `152.0` |
| Selenium Firefox | `firefox` | `154.0` |
| Selenium Edge | `msedge` | `151.0` |
| Selenium headless CI | `chrome` / `firefox` / `msedge` | `*-min` |
| `@playwright/test` в CI | `playwright-chromium` | `1.62.1` |
| Playwright + Firefox | `playwright-firefox` | `1.62.1` |
| Playwright + Edge | `playwright-msedge` | `1.62.1` |
| Playwright + Safari-движок | `playwright-webkit` | `1.62.1` |
