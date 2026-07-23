# Release v3.0.0 — qa-guru/selenoid

**Дата:** 23 июля 2026  
**Предыдущий:** [v2.3.0](RELEASE_v2.3.0.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v3.0.0  
**Stack cut:** hub + cm → **v3.0.0**; UI — отдельная линия **v3.0.x** на `selenoid-ui` `main`.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Selenoid 3 hub line** | Prod hub semver переходит на **v3.0.0** с `main`; pin-ветки 2.x заморожены |
| **Playwright accessKey** | Публичный stand: обязательный `accessKey` для Playwright WS (nginx + hub) |
| **Video list API** | Paginated `/video` list с limit/offset |
| **Android catalog** | SSOT: android **16.0** (`qaguru/android:16`) + legacy `selenoid/android:4.4`, `10.0` |
| **Docker API 1.55** | Moby client + Engine **29.x** (наследие v2.3.0 baseline) |

Browser-image теги — без изменений политики; prod overlay — `browsers-production.json`.

---

## Миграция с v2.3.0

1. Обновить hub binary до `v3.0.0`.
2. Обновить cm до `v3.0.0`.
3. UI — **v3.0.x** (уже на prod); hub/cm pin **v2.3.0** больше не использовать.

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v3.0.0/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
./selenoid -conf /etc/selenoid/browsers.json -limit 25
```

Docker: `docker pull qaguru/selenoid:v3.0.0`

Связанные: [cm v3.0.0](https://github.com/qa-guru/cm/releases/tag/v3.0.0), [selenoid-ui v3.0.7+](https://github.com/qa-guru/selenoid-ui/releases).

---

## Cut checklist

1. `main` green (`go test ./...`, CI build).
2. `git tag -a v3.0.0 -m "v3.0.0"` → push tag → GitHub Release (published).
3. Prod deploy: `SELENOID_VERSION=v3.0.0`, `CM_VERSION=v3.0.0`, `SELENOID_UI_VERSION=v3.0.7+`.
