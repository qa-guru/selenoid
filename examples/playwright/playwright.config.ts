import { defineConfig } from '@playwright/test';
import dotenv from 'dotenv';

dotenv.config();

const selenoidOptions: Record<string, string> = {
  name: 'playwright-smoke',
  sessionTimeout: '5m',
  enableVideo: 'true',
  enableVNC: 'true',
};

function buildPlaywrightWsEndpoint(): string {
  const fromEnv =
    process.env.PW_TEST_CONNECT_WS_ENDPOINT ??
    process.env.PLAYWRIGHT_WS_ENDPOINT;

  if (fromEnv?.includes('?')) {
    return fromEnv;
  }

  const base =
    fromEnv ?? 'ws://localhost:4444/playwright/playwright-chromium/1.61.1';

  return `${base}?${new URLSearchParams(selenoidOptions)}`;
}

const playwrightWsEndpoint = buildPlaywrightWsEndpoint();

export default defineConfig({
  testDir: './tests',
  retries: 0,
  workers: 1,
  use: {
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    connectOptions: {
      wsEndpoint: playwrightWsEndpoint,
    },
  },
  projects: [
    {
      name: 'chromium',
      use: { browserName: 'chromium' },
    },
  ],
});
