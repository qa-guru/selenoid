# Release v2.0.9 — qa-guru/selenoid

**Дата:** 27 июня 2026  
**Предыдущий:** [v2.0.8](RELEASE_v2.0.8.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.0.9

Документационный релиз: без изменений hub-кода, обновлены README и Playwright-доки.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **README** | Обзор экосистемы qa-guru, таблица `browsers.json`, toolchain (Go 1.23, Docker API 1.45) |
| **docs/playwright.md** | Русская документация; убран устаревший раздел basic auth |
| **examples/playwright** | README на русском |
| **RELEASE_v2.0.0** | Ссылки на отдельные репозитории вместо монорепо |

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.0.9/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 20
```
