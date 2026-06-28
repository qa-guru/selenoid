# Release v2.1.1 — qa-guru/selenoid

**Дата:** 28 июня 2026  
**Предыдущий:** [v2.1.0](RELEASE_v2.1.0.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v2.1.1

Патч-релиз: Playwright-примеры и документация с **`enableVNC=true&enableVideo=true`** по умолчанию.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Playwright URL** | README, `docs/playwright.md`, `docs/browser-versions.md` — query-параметры VNC и video в endpoint |
| **examples/playwright** | `.env.example` с корректным путём `playwright-chromium`; `playwright.config.ts` добавляет selenoid options, если их нет в env |

---

## Установка / обновление

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v2.1.1/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
DOCKER_API_VERSION=1.45 ./selenoid -conf /etc/selenoid/browsers.json -limit 20
```
