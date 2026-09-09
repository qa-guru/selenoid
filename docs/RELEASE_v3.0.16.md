# Release v3.0.16 — qa-guru/selenoid

**Дата:** 9 сентября 2026  
**Предыдущий:** [v3.0.15](https://github.com/qa-guru/selenoid/releases/tag/v3.0.15)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v3.0.16  
**Stack cut:** hub → **v3.0.16**; cm → **v3.0.5**; UI → **v3.0.58** (cut separately).

## Что нового

| Изменение | Описание |
|-----------|----------|
| **-min caps** | `*-min` images are headless CI. `enableVNC` / `enableVideo` / `enableHAR` now fail **before** the container starts: HTTP **400**, W3C `invalid argument`, message names the tag and the flags. Not a ChromeDriver 500. |
| **Playwright Stop** | DELETE no longer waits up to 35s for CDP HAR attach. Min images have no `:7070`; Stop returns immediately and still flushes HAR if attach already finished. |

```bash
docker pull qaguru/selenoid:v3.0.16
```

Связанные: [selenoid-ui v3.0.58](https://github.com/qa-guru/selenoid-ui/releases/tag/v3.0.58).
