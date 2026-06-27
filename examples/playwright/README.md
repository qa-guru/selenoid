# Пример Playwright

Smoke-тест нативной поддержки Playwright в Selenoid.

## Предварительные условия

```bash
go build -o selenoid .
docker pull qaguru/playwright-chromium:1.61.1
./selenoid -conf config/browsers.json -limit 5
```

## Запуск

```bash
npm install
cp .env.example .env
npm test
```

Подробности endpoint — в [docs/playwright.md](../../docs/playwright.md).
