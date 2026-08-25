# Release v3.0.14 — qa-guru/selenoid

**Дата:** 25 августа 2026  
**Предыдущий:** [v3.0.13](https://github.com/qa-guru/selenoid/releases/tag/v3.0.13)  
**GitHub:** https://github.com/qa-guru/selenoid/releases/tag/v3.0.14  
**Stack cut:** hub → **v3.0.14**; cm → **v3.0.3**; UI → **v3.0.51**.

## Что нового

| Изменение | Описание |
|-----------|----------|
| **Playwright HAR on Stop** | Slot is reserved before async CDP attach. `DELETE` / WS close waits for the recorder, writes `/har/<id>.har`, then kills the container. Empty HAR is still persisted so the session stays in `/sessions/` after reload. |
| **Playwright logs** | `enableLog` copies container stdout/stderr on cancel (same as WebDriver) and renames to `<id>.log`. |
| **Headed VNC window** | After HAR `/json/new` bootstrap, set Chromium window bounds to `screenResolution` so VNC is not stuck at the default Xvfb size. |

```bash
docker pull qaguru/selenoid:v3.0.14
```

Связанные: [selenoid-ui v3.0.51](https://github.com/qa-guru/selenoid-ui/releases/tag/v3.0.51).
