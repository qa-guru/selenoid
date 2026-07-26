# Release v3.0.2 — qa-guru/selenoid

**Дата:** 26 июля 2026  
**Предыдущий:** [v3.0.1](RELEASE_v3.0.1.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v3.0.2  
**Stack cut:** hub → **v3.0.2**; cm — **v3.0.1** (без изменений); UI — **v3.0.8**.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **CLI rename** | Flag `-playwright-access-key` → **`-access-key`**; var `playwrightAccessKeys` → `accessKeys` |
| **Query param** | Без изменений: обязательный `?accessKey=` (alias `access_key`) при непустом `-access-key` |
| **No dual-flag** | Legacy имя flag убрано; SSOT на edge — nginx map `?accessKey=` (prod flag не передаётся) |

Выравнивание имени с UI guest creds (`VITE_HUB_ACCESS_KEY` / hubAuth) и docs selenoid-ui.

---

## Обновление с v3.0.1

1. Обновить hub binary до `v3.0.2`.
2. Если где-то передавали `-playwright-access-key` — заменить на `-access-key` (на prod.qa.guru flag не используется).
3. cm / UI — без обязательного bump.

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v3.0.2/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
./selenoid -conf /etc/selenoid/browsers.json -limit 25 -har-output-dir /opt/selenoid/har
```

Docker: `docker pull qaguru/selenoid:v3.0.2`

Связанные: [cm v3.0.1](https://github.com/qa-guru/cm/releases/tag/v3.0.1), [selenoid-ui v3.0.8](https://github.com/qa-guru/selenoid-ui/releases/tag/v3.0.8).

---

## Cut checklist

1. `main` green (`go test ./...`, CI build).
2. `git tag -a v3.0.2 -m "v3.0.2"` → push tag → GitHub Release (published) → `release.yml` assets + `qaguru/selenoid:v3.0.2`.
3. Prod deploy (отдельный чат): `SELENOID_VERSION=v3.0.2`, `CM_VERSION=v3.0.1`, `SELENOID_UI_VERSION=v3.0.8`.
