# Release v3.0.3 — qa-guru/selenoid

**Дата:** 28 июля 2026  
**Предыдущий:** [v3.0.2](RELEASE_v3.0.2.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v3.0.3  
**Stack cut:** hub → **v3.0.3**; cm — **v3.0.1** (без изменений); UI — **v3.0.12**.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Playwright DevTools :7070** | Hub всегда публикует контейнерный порт `7070` для Playwright-сессий (как warm WebDriver) — hub-HAR / `se:cdp` / `/devtools/<id>/` |
| **Playwright hub-HAR retry** | `enableHAR=true` стартует capture асинхронно с retry attach к `/page` после client `newPage()` (launchServer изначально без page) |
| **`logName` query** | Парсинг `logName` из Playwright session query params |
| **`socksProxy` → `PW_PROXY`** | Query `socksProxy` пробрасывается в env браузера при launch |

Требует Chromium-family образов с DevTools proxy: `qaguru/playwright-chromium|chrome|msedge:1.61.1` (firefox / webkit / `*-min` — без `:7070` by design).

---

## Обновление с v3.0.2

1. Обновить hub binary до `v3.0.3`.
2. Убедиться, что на хосте уже есть образы с `:7070` (`docker pull qaguru/playwright-chromium:1.61.1` и т.п.).
3. cm / UI — без обязательного bump (prod UI остаётся **v3.0.12**).

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v3.0.3/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
./selenoid -conf /etc/selenoid/browsers.json -limit 25 -har-output-dir /opt/selenoid/har
```

Docker: `docker pull qaguru/selenoid:v3.0.3`

Связанные: [cm v3.0.1](https://github.com/qa-guru/cm/releases/tag/v3.0.1), [selenoid-ui v3.0.12](https://github.com/qa-guru/selenoid-ui/releases/tag/v3.0.12), [browser-image playwright :7070](https://github.com/qa-guru/browser-image).

---

## Cut checklist

1. `main` green (`go test ./...`, CI build).
2. `git tag -a v3.0.3 -m "v3.0.3"` → push tag → GitHub Release (published) → `release.yml` assets + `qaguru/selenoid:v3.0.3`.
3. Prod deploy: `SELENOID_VERSION=v3.0.3`, `CM_VERSION=v3.0.1`, `SELENOID_UI_VERSION=v3.0.12`.
4. Smoke: Playwright `enableHAR` → `GET /har/<id>.har` 2xx + `log.entries`; UI HarViewer.
