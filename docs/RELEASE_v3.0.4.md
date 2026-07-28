# Release v3.0.4 — qa-guru/selenoid

**Дата:** 28 июля 2026  
**Предыдущий:** [v3.0.3](RELEASE_v3.0.3.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v3.0.4  
**Stack cut:** hub → **v3.0.4**; cm — **v3.0.1** (без изменений); UI — **v3.0.13** (`harContent` wire).

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **`harContent` meta\|bodies** | Capability поверх `enableHAR`: default **`meta`** (как v3.0.3 — URL/headers/status/size/mimeType, без `content.text`); opt-in **`bodies`** — best-effort `Network.getResponseBody` → `content.text` (+ `encoding` при base64) |
| **PW query** | `?enableHAR=true&harContent=bodies` (omit / `meta` = default) |
| **One writer** | По-прежнему запрет dual-writer: не сочетать hub `enableHAR` с client `recordHar` / `HarCapture` на одной сессии |
| **Not ≡ recordHar** | Даже `bodies` не claim полной parity со status/size/text Playwright `recordHar` |

Требует warm Chrome / Chromium-family с DevTools `:7070` (как v3.0.3). firefox / webkit / `*-min` / Android — без HAR bodies claim.

---

## Обновление с v3.0.3

1. Обновить hub binary до `v3.0.4`.
2. UI — **v3.0.13** (Capabilities select `harContent`, только при `enableHAR`).
3. Default path без изменений: `enableHAR` без `harContent` = meta.

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v3.0.4/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
./selenoid -conf /etc/selenoid/browsers.json -limit 25 -har-output-dir /opt/selenoid/har
```

Docker: `docker pull qaguru/selenoid:v3.0.4`

Связанные: [cm v3.0.1](https://github.com/qa-guru/cm/releases/tag/v3.0.1), [selenoid-ui v3.0.13](https://github.com/qa-guru/selenoid-ui/releases/tag/v3.0.13), ADR [`009-har-content-meta-bodies`](../../../../docs/adr/009-har-content-meta-bodies.md).

---

## Cut checklist

1. `main` green (`go test ./...`, CI build).
2. `git tag -a v3.0.4 -m "v3.0.4"` → push tag → GitHub Release (published) → `release.yml` assets + `qaguru/selenoid:v3.0.4`.
3. Prod deploy: `SELENOID_VERSION=v3.0.4`, `CM_VERSION=v3.0.1`, `SELENOID_UI_VERSION=v3.0.13`.
4. Smoke: WD/PW `enableHAR` meta → `withContentText==0`; `harContent=bodies` → `withContentText>=1` (best-effort); UI Capabilities + HarViewer.
