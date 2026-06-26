# Playwright example

Smoke test for native Playwright support in Selenoid.

## Prerequisites

```bash
go build -o selenoid .
docker pull qaguru/playwright-chromium:1.61.1
./selenoid -conf config/browsers.json -limit 5
```

## Run

```bash
npm install
cp .env.example .env
npm test
```

See [docs/playwright.md](../../docs/playwright.md) for endpoint details.
