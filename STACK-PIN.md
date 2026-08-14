# Stack pin: main / v3.0.x (Selenoid 3)

**Репозиторий:** Hub / engine (qa-guru/selenoid)

Этот файл на **`main`** описывает живой hub toolchain. Pin-ветки 2.x — отдельные frozen `STACK-PIN.md`.

| Поле | Значение |
|------|----------|
| Линия | Selenoid 3 |
| Stack semver | hub cut **v3.0.10** (prod pin [selenoid.qa.guru](https://selenoid.qa.guru)) |
| Docker API | TBD (paired с prod Engine 29.6+) |
| Docker Engine | TBD (prod: Debian 12 · Docker 29.6) |
| Go | 1.26.5+ |
| Go (примечание) | Факт `go.mod` + `toolchain go1.26.5` |
| Prod | [selenoid.qa.guru](https://selenoid.qa.guru) |
| Git anchor | `main` |
| Docker image | `qaguru/selenoid:v3.0.x` |
| **Фичи v3+** | native Playwright WS, HAR (`harContent`), warm-pool metrics + Chrome WD hub-attach |

## Selenoid 2 maintenance pin (не путать)

Maintenance **v2.3.0** / React **18** — только pin-ветка
[`selenoid2-1.55-engine29.6-go1.26-react18`](https://github.com/qa-guru/selenoid/tree/selenoid2-1.55-engine29.6-go1.26-react18)
(`qaguru/selenoid:v2.3.0`, Docker API **1.55**). Rollback **v2.2.1** / React 16 —
[`selenoid2-1.45-engine26.1-go1.26-react16`](https://github.com/qa-guru/selenoid/tree/selenoid2-1.45-engine26.1-go1.26-react16).

См. также: [`projects/selenoid-home/README.md`](https://github.com/qa-guru/zero-design-system/blob/master/projects/selenoid-home/README.md) (monorepo SSOT).
