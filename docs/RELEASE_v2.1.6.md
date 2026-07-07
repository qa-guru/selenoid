# Release v2.1.6 — qa-guru/selenoid

**Дата:** 7 июля 2026  
**Предыдущий:** [v2.1.5](RELEASE_v2.1.5.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.1.6

Патч-релиз стека **v2.1.6**: выравнивание с `selenoid-ui` v2.1.6, Playwright **1.61.1** в дефолтном каталоге, тесты и docs.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **browsers.json** | Playwright default **1.61.1** (+ `-min`); `shmSize`, `hosts` для chrome/webdriver и PW nodes |
| **Тесты** | Playwright WS/docker unit coverage (`playwright_test`, `playwright_docker_*`, `service_find_playwright`) |
| **Документация** | `browser-versions.md`, `playwright.md`, driver catalog (`browser-driver-catalog.xlsx` + gen script) |

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.1.6/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 20
```

Docker: `docker pull qaguru/selenoid:v2.1.6`
