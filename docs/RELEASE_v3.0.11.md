# Release v3.0.11 — qa-guru/selenoid

**Дата:** 14 августа 2026  
**Предыдущий:** [v3.0.9](https://github.com/qa-guru/selenoid/releases/tag/v3.0.9)  
**Stack cut:** hub → **v3.0.11**; cm → **v3.0.2**; UI → **v3.0.31**.

## Что нового

| Изменение | Описание |
|-----------|----------|
| **hotReady / hotTotal** | Hub `/status` splits orchestrator `/pool/slots` by `pool=hot`. Warm stays `pool!=hot`. Empty `-warm-pool-url` → `0/0` for both. |
| **Go** | toolchain **1.26.6** + `golang.org/x/net` v0.55.0 — govulncheck (stdlib + idna), same as UI v3.0.30 |
| **Docs** | Cold / warm 4/4 / hot named in hub-attach; PW hub-attach stays cold. |

```bash
docker pull qaguru/selenoid:v3.0.11
```
