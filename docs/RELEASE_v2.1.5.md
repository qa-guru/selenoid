# Release v2.1.5 — qa-guru/selenoid

**Дата:** 2 июля 2026  
**Предыдущий:** [v2.1.1](RELEASE_v2.1.1.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.1.5

Патч-релиз стека **v2.1.5**: выравнивание с `selenoid-ui` и `cm`, активный каталог только на `qaguru/*`, документация политики версий.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **browsers.json** | Активный стек: `qaguru/webdriver-chrome:148` (+ `148-min`), Playwright **1.61.1** (+ `-min`); legacy `twilio/*`, Firefox/Edge WebDriver убраны из дефолтного конфига |
| **Документация** | [browser-versions.md](browser-versions.md) — политика трёх слоёв (hub / PW nodes / WD nodes); новый [driver-versions-catalog.md](driver-versions-catalog.md) |
| **Min-образы** | `playwright-chromium:1.61.1-min`, `webdriver-chrome:148-min` в каталоге |
| **Тесты** | Go unit: readable display names через `t.Run` для Allure/TestOps |

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.1.5/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 20
```

Docker: `docker pull qaguru/selenoid:v2.1.5`
