# Release v3.0.13 — qa-guru/selenoid

**Дата:** 17 августа 2026  
**Предыдущий:** [v3.0.12](https://github.com/qa-guru/selenoid/releases/tag/v3.0.12)  
**Stack cut:** hub → **v3.0.13**; cm → **v3.0.3**; UI → **v3.0.36**.

## Что нового

| Изменение | Описание |
|-----------|----------|
| **`-pool-url`** | Alias for `-warm-pool-url`. Canonical flag stays `-warm-pool-url` (prod unit / cm). |
| **Docs** | Container-reuse naming; sidecar repo [qa-guru/selenoid-pool](https://github.com/qa-guru/selenoid-pool). |

Pool compose / hot lease live on box1 are unchanged. This cut does not stop warm/hot slots.

```bash
docker pull qaguru/selenoid:v3.0.13
```

Связанные: [cm v3.0.3](https://github.com/qa-guru/cm/releases/tag/v3.0.3), [selenoid-ui v3.0.36](https://github.com/qa-guru/selenoid-ui/releases/tag/v3.0.36).
