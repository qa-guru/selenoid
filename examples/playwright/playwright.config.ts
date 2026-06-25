import { defineConfig } from '@playwright/test';
import dotenv from 'dotenv';

dotenv.config();

const playwrightWsEndpoint =
  process.env.PW_TEST_CONNECT_WS_ENDPOINT ??
  process.env.PLAYWRIGHT_WS_ENDPOINT;

if (!playwrightWsEndpoint) {
  throw new Error('Set PLAYWRIGHT_WS_ENDPOINT or PW_TEST_CONNECT_WS_ENDPOINT');
}

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
