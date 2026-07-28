# Release v3.0.5 — qa-guru/selenoid

**Дата:** 28 июля 2026  
**Предыдущий:** [v3.0.4](RELEASE_v3.0.4.md)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v3.0.5  
**Stack cut:** hub → **v3.0.5**; cm → **v3.0.2**; UI → **v3.0.14**.

---

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Go module path** | `github.com/aerokube/selenoid` → **`github.com/qa-guru/selenoid`** — imports, CI coverpkg, docs |
| **HAR bodies fix** | `content.size` из decoded body в режиме `harContent=bodies` |

Runtime / Docker API / capabilities без breaking changes для потребителей hub (WebDriver, Playwright WS).

---

## Обновление с v3.0.4

1. Обновить hub binary до `v3.0.5` (или Docker `qaguru/selenoid:v3.0.5`).
2. Go-импортёрам: заменить import path на `github.com/qa-guru/selenoid`.

```bash
curl -sL https://github.com/qa-guru/selenoid/releases/download/v3.0.5/selenoid_linux_amd64 -o selenoid
chmod +x selenoid
```
