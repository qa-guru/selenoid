# Каталог версий драйверов и browser nodes

Одна таблица: hub, browser nodes, legacy. Политика: [browser-versions.md](browser-versions.md). Сверка: июль 2026.

Конфиги: dev/cm → PW **1.60.0**, WD **148.0** · CI `tests-java` → PW **1.61.1**, WD **148.0**.

| Версия / компонент | Браузер | Статус | Microsoft | aerokube | Twilio | qaguru | cm |
|--------------------|---------|--------|-----------|----------|--------|--------|-----|
| Hub | — | active | aerokube/selenoid v1.11.x | `aerokube/selenoid:latest-release` | — | `qaguru/selenoid:latest-release` | — |
| UI | — | active | aerokube/selenoid-ui | `aerokube/selenoid-ui` | — | `qaguru/selenoid-ui:latest-release` | — |
| Video | — | active | — | `selenoid/video-recorder:latest-release` legacy | — | `qaguru/video-recorder:latest` | — |
| PW 1.60.0 | Chromium | active | 148.0.7778.96; MCR v1.60.0-noble | — | — | `playwright-chromium:1.60.0` | — |
| PW 1.60.0 | Firefox | active | 150.0.2 | — | — | `playwright-firefox:1.60.0` | — |
| PW 1.60.0 | WebKit | active | 26.4 | — | — | `playwright-webkit:1.60.0` | — |
| PW 1.60.0 | Chrome stable | active | stable channel | — | — | `playwright-chrome:1.60.0` | — |
| PW 1.60.0 | MS Edge | active | stable channel | — | — | `playwright-msedge:1.60.0` | — |
| PW 1.60.0 | Opera | active | — | `selenoid/opera:*` legacy | — | — | operadriver 106.0.5249.119 |
| PW 1.60.0 | WebDriver Chrome | active | CfT 148.0.7778.96 | `selenoid/chrome:148.x` legacy | `chrome_stable_148` legacy | `webdriver-chrome:148` | chromedriver 148.0.7778.178 |
| PW 1.60.0-min | Chromium | active | 148.0.7778.96 headless | — | — | `playwright-chromium:1.60.0-min` | — |
| PW 1.60.0-min | WebDriver Chrome | active | CfT 148.0.7778.96 | — | — | `webdriver-chrome:148-min` | — |
| PW 1.61.0 | Chromium | published | 149.0.7827.55; MCR v1.61.0-noble | — | — | `playwright-chromium:1.61.0` | — |
| PW 1.61.0 | Firefox | published | 151.0 | — | — | `playwright-firefox:1.61.0` | — |
| PW 1.61.0 | WebKit | published | 26.5 | — | — | `playwright-webkit:1.61.0` | — |
| PW 1.61.0 | Chrome stable | published | stable channel | — | — | `playwright-chrome:1.61.0` | — |
| PW 1.61.0 | MS Edge | published | stable channel | — | — | `playwright-msedge:1.61.0` | — |
| PW 1.61.0 | WebDriver Chrome | published | CfT 149.0.7827.55 | — | `chrome_stable_149` legacy | — | — |
| PW 1.61.1 | Chromium | active (CI) | 149.0.7827.55; MCR v1.61.1-noble | — | — | `playwright-chromium:1.61.1` | — |
| PW 1.61.1 | Firefox | published | 151.0 | — | — | `playwright-firefox:1.61.1` | — |
| PW 1.61.1 | WebKit | published | 26.5 | — | — | `playwright-webkit:1.61.1` | — |
| PW 1.61.1 | Chrome stable | published | stable channel | — | — | `playwright-chrome:1.61.1` | — |
| PW 1.61.1 | MS Edge | published | stable channel | — | — | `playwright-msedge:1.61.1` | — |
| PW 1.61.1-min | Chromium | active (CI) | 149.0.7827.55 headless | — | — | `playwright-chromium:1.61.1-min` | — |
| PW 1.61.1-min | WebDriver Chrome | active (CI) | CfT 149.0.7827.55 | — | — | `webdriver-chrome:149-min` | — |
| WD chrome 148.0 | Chromium | active | CfT 148.0.7778.96 | — | — | — | — |
| WD chrome 148.0 | Firefox | active | — | `selenoid/firefox:*` legacy | `firefox_stable_148` legacy | — | geckodriver 0.37.0 |
| WD chrome 148.0 | MS Edge | active | — | — | — | — | msedgedriver 145.0.3800.97 |
| WD chrome 148.0 | Opera | active | — | `selenoid/opera:*` legacy | — | — | operadriver 106.0.5249.119 |
| WD chrome 148.0 | WebDriver Chrome | active | CfT 148.0.7778.96 + chromedriver | `selenoid/chrome:148.x` legacy | `chrome_stable_148` legacy | `webdriver-chrome:148` warm | chromedriver 148.0.7778.178 |
| WD chrome 148.0-min | WebDriver Chrome | active | CfT 148.0.7778.96 | — | — | `webdriver-chrome:148-min` | — |
| WD legacy | Chromium | legacy | CfT 146–147 | `selenoid/chrome:146.x`…`147.x` | `chrome_stable_146`…`_147` | не публикуем | — |
| WD legacy | Firefox | legacy | — | `selenoid/firefox:48.0`…`120.0` | `firefox_stable_148`…`_150` | → `playwright-firefox` | geckodriver 0.37.0 |
| WD legacy | MS Edge | legacy | — | — | — | → `playwright-msedge` | msedgedriver 145.0.3800.97 |
| WD legacy | Opera | legacy | — | `selenoid/opera:*` | — | — | operadriver 106.0.5249.119 |
| WD legacy | WebDriver Chrome | legacy | CfT 146–147 | `selenoid/chrome:146.x`…`147.x` | `chrome_stable_146`…`_147` | — | — |

**Браузер** = столбец в hub: `playwright-chromium` → Chromium, `chrome` → WebDriver Chrome, и т.д. Движки PW — [Microsoft `browsers.json`](https://github.com/microsoft/playwright/blob/main/packages/playwright-core/browsers.json). `chrome:148.0` ≠ `playwright-chromium:1.60.0` (разный протокол).

**docker pull (active):** `qaguru/webdriver-chrome:148` `:148-min` `:149-min` · `qaguru/playwright-chromium:1.60.0` `:1.60.0-min` `:1.61.1` `:1.61.1-min` · `playwright-firefox` / `webkit` / `chrome` / `msedge` `:1.60.0` · `qaguru/video-recorder:latest`.

См. [browser-versions.md](browser-versions.md) · [playwright.md](playwright.md).
