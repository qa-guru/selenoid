# Release v3.0.1 — qa-guru/selenoid

**Дата:** 26 июля 2026  
**Предыдущий:** [v3.0.0](RELEASE_v3.0.0.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v3.0.1  
**Stack cut:** hub + cm → **v3.0.1**; UI — **v3.0.8** на `selenoid-ui` `main`.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Finished-session metadata** | Всегда persist name/quota/started/finished (без metadata build tag); quota user в метаданных |
| **`/sessions/?json` enrich** | Группировка video/log/har по session id; meta из metadata (mtime fallback) для Finished sessions UI |
| **Native HAR** | CDP HAR capture: `-har-output-dir`, `/har/` list + download |
| **Timestamps omitempty** | `started`/`finished` как `*time.Time` — нулевые значения не утекают как `0001-01-01` в UI |

Browser-image теги — без изменений политики; prod overlay — `browsers-production.json`.

---

## Обновление с v3.0.0

1. Обновить hub binary до `v3.0.1`.
2. Обновить cm до `v3.0.1` (deps + android embed tests).
3. UI — **v3.0.8** (Finished sessions + HAR viewer).

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v3.0.1/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
./selenoid -conf /etc/selenoid/browsers.json -limit 25 -har-output-dir /opt/selenoid/har
```

Docker: `docker pull qaguru/selenoid:v3.0.1`

Связанные: [cm v3.0.1](https://github.com/qa-guru/cm/releases/tag/v3.0.1), [selenoid-ui v3.0.8](https://github.com/qa-guru/selenoid-ui/releases/tag/v3.0.8).

---

## Cut checklist

1. `main` green (`go test ./...`, CI build).
2. `git tag -a v3.0.1 -m "v3.0.1"` → push tag → GitHub Release (published) → `release.yml` assets + `qaguru/selenoid:v3.0.1`.
3. Prod deploy (отдельный чат): `SELENOID_VERSION=v3.0.1`, `CM_VERSION=v3.0.1`, `SELENOID_UI_VERSION=v3.0.8`.
