import { test, expect } from '@playwright/test';

test('opens example.com on remote Chromium', async ({ page }) => {
  await page.goto('https://example.com');
  await expect(page).toHaveTitle(/Example Domain/);
  await expect(page.locator('h1')).toHaveText('Example Domain');
});
