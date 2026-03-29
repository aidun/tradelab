import { chromium } from "@playwright/test";
import fs from "node:fs/promises";
import path from "node:path";

const docsScreenshotDir = path.resolve(process.cwd(), "..", "docs", "screenshots");
const appBaseUrl = process.env.DOCS_SCREENSHOT_BASE_URL ?? "http://127.0.0.1:3000";

const responses = [
  {
    orders: [
      {
        id: "order-1",
        walletID: "wallet-1",
        marketSymbol: "XRP/USDT",
        quoteAmount: 50,
        expectedPrice: 0.67,
        status: "filled",
        createdAt: "2026-03-29T12:00:00Z"
      }
    ],
    activity: [
      {
        id: "log-1",
        walletID: "wallet-1",
        logType: "trade",
        title: "Demo buy recorded",
        message: "A demo market buy was created for XRP/USDT.",
        createdAt: "2026-03-29T12:01:00Z"
      }
    ],
    portfolio: {
      walletID: "wallet-1",
      baseCurrency: "USDT",
      totalValue: 10000,
      cashBalance: 9950,
      balances: [{ walletID: "wallet-1", assetSymbol: "USDT", available: 9950 }],
      positions: []
    }
  },
  {
    orders: [
      {
        id: "order-2",
        walletID: "wallet-1",
        marketSymbol: "XRP/USDT",
        quoteAmount: 75,
        expectedPrice: 0.69,
        status: "filled",
        createdAt: "2026-03-29T12:02:00Z"
      }
    ],
    activity: [
      {
        id: "log-2",
        walletID: "wallet-1",
        logType: "trade",
        title: "Demo buy executed",
        message: "A second demo market buy was created for XRP/USDT.",
        createdAt: "2026-03-29T12:03:00Z"
      }
    ],
    portfolio: {
      walletID: "wallet-1",
      baseCurrency: "USDT",
      totalValue: 10025,
      cashBalance: 9875,
      balances: [{ walletID: "wallet-1", assetSymbol: "USDT", available: 9875 }],
      positions: [
        {
          id: "position-1",
          userID: "user-1",
          walletID: "wallet-1",
          marketID: "market-1",
          marketSymbol: "XRP/USDT",
          baseAsset: "XRP",
          quoteAsset: "USDT",
          status: "open",
          entryQuantity: 108.7,
          entryPriceAvg: 0.69,
          currentPrice: 0.69,
          positionValue: 75,
          unrealizedPnL: 0,
          openedAt: "2026-03-29T12:02:00Z"
        }
      ]
    }
  }
];

async function installRoutes(page) {
  let orderStateIndex = 0;

  await page.addInitScript(() => {
    window.localStorage.clear();
  });

  await page.route("**/api/v1/**", async (route) => {
    const request = route.request();
    const url = request.url();

    if (url.endsWith("/api/v1/sessions/demo")) {
      await route.fulfill({
        status: 201,
        contentType: "application/json",
        body: JSON.stringify({
          session: {
            id: "session-1",
            user_id: "user-1",
            wallet_id: "wallet-1",
            token: "token-1",
            expires_at: "2026-04-29T12:00:00Z"
          }
        })
      });
      return;
    }

    if (url.includes("/candles")) {
      const interval = url.includes("interval=15m") ? "15m" : "1h";
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          candles: [
            {
              openTime: "2026-03-29T10:00:00Z",
              closeTime: "2026-03-29T10:59:59Z",
              openPrice: 0.62,
              highPrice: 0.64,
              lowPrice: 0.61,
              closePrice: 0.63,
              baseVolume: 1200000,
              quoteVolume: 756000,
              trades: 8000
            },
            {
              openTime: "2026-03-29T11:00:00Z",
              closeTime: "2026-03-29T11:59:59Z",
              openPrice: 0.63,
              highPrice: 0.65,
              lowPrice: 0.62,
              closePrice: 0.64,
              baseVolume: 1400000,
              quoteVolume: 896000,
              trades: 9200
            }
          ],
          meta: {
            source: interval === "15m" ? "stale" : "fresh",
            generated_at: "2026-03-29T12:05:00Z"
          }
        })
      });
      return;
    }

    if (url.endsWith("/api/v1/markets")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          markets: [
            {
              id: "market-1",
              symbol: "XRP/USDT",
              baseAsset: "XRP",
              quoteAsset: "USDT",
              minNotional: 10,
              exchange: "demo"
            }
          ]
        })
      });
      return;
    }

    if (url.endsWith("/api/v1/orders") && request.method() === "POST") {
      orderStateIndex = 1;
      await route.fulfill({
        status: 201,
        contentType: "application/json",
        body: JSON.stringify({ order: { id: "order-2" } })
      });
      return;
    }

    if (url.endsWith("/api/v1/orders")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ orders: responses[orderStateIndex].orders })
      });
      return;
    }

    if (url.endsWith("/api/v1/activity")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ activity: responses[orderStateIndex].activity })
      });
      return;
    }

    if (url.includes("/api/v1/portfolios/")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ portfolio: responses[orderStateIndex].portfolio })
      });
      return;
    }

    await route.fallback();
  });
}

async function capture() {
  await fs.mkdir(docsScreenshotDir, { recursive: true });

  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1600, height: 2400 } });

  await installRoutes(page);
  await page.goto(appBaseUrl);
  await page.getByRole("heading", { name: /demo execution with/i }).waitFor();
  await page.screenshot({
    path: path.join(docsScreenshotDir, "dashboard-overview.png"),
    fullPage: true
  });

  await page.getByRole("button", { name: "15m" }).click();
  await page.getByText(/feed stale fallback/i).waitFor();
  await page.screenshot({
    path: path.join(docsScreenshotDir, "chart-stale-feed.png"),
    fullPage: true
  });

  await page.getByRole("button", { name: /run demo buy/i }).click();
  await page.getByText(/demo buy executed for xrp\/usdt/i).waitFor();
  await page.screenshot({
    path: path.join(docsScreenshotDir, "demo-buy-success.png"),
    fullPage: true
  });

  await browser.close();
}

capture().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
