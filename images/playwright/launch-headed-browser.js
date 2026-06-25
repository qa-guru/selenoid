const { chromium, firefox, webkit } = require("playwright");

const browserTypes = { chromium, firefox, webkit };

const pwBrowser = process.env.PW_BROWSER || "chromium";
const port = process.env.PW_PORT || "3000";
const browserType = browserTypes[pwBrowser] || chromium;

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

(async () => {
  for (let attempt = 0; attempt < 60; attempt++) {
    try {
      const browser = await browserType.connect({
        wsEndpoint: `ws://127.0.0.1:${port}/`,
        headers: {
          "x-playwright-browser": pwBrowser,
          "x-playwright-launch-options": JSON.stringify({ headless: false }),
        },
      });
      const page = await browser.newPage();
      await page.goto("about:blank");
      browser.on("disconnected", () => process.exit(0));
      return;
    } catch (err) {
      await sleep(500);
      if (attempt === 59) {
        console.error("connect attempt failed:", err.message || err);
      }
    }
  }
  console.error("Failed to launch headed browser for VNC");
  process.exit(1);
})();
