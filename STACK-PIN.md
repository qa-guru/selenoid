# Stack pin: selenoid2-1.55-engine29.6-go1.26-react18

**Репозиторий:** Hub / engine (qa-guru/selenoid)

| Поле | Значение |
|------|----------|
| Линия | Selenoid 2 (maintenance) |
| Stack semver | v2.3.0 (in progress) |
| Docker API | 1.55 |
| Docker Engine | 29.x (рекоменд. 29.6+) |
| Go | 1.26.5 |
| Go (примечание) | Целевой toolchain cut v2.3.0; часть файлов на main может ещё показывать 1.45 до завершения cut |
| React | 18 |
| UI | Vite 6 |
| Prod reference | selenoid.autotests.cloud остаётся на v2.2.x до отдельного cut |
| До | До фазы 3 / Selenoid 3 UI (selenoid.qa.guru) |
| Git anchor | `main` (pre-cut v2.3.0) |
| Docker image | `qaguru/selenoid:v2.3.0` (target) |
| Moby client | `moby/moby/api v1.55.0` (target) |
| React (stack pin) | 18 — paired [selenoid-ui](https://github.com/qa-guru/selenoid-ui) |

См. также: [`projects/selenoid-home/README.md`](https://github.com/qa-guru/zero-design-system/blob/master/projects/selenoid-home/README.md) (monorepo SSOT).
