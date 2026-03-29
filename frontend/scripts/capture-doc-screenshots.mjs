import { chromium } from "@playwright/test";
import fs from "node:fs/promises";
import path from "node:path";

const docsScreenshotDir = path.resolve(process.cwd(), "..", "docs", "screenshots");
const appBaseUrl = process.env.DOCS_SCREENSHOT_BASE_URL ?? "http://127.0.0.1:3000";

function state(index) {
  const states = [
    {
      portfolio: {
        walletID: "wallet-1",
        baseCurrency: "USDT",
        accountingMode: "average_cost",
        totalValue: 10000,
        cashBalance: 9950,
        positionValue: 50,
        realizedPnL: 0,
        unrealizedPnL: 0,
        balances: [
          { walletID: "wallet-1", assetSymbol: "USDT", available: 9950 },
          { walletID: "wallet-1", assetSymbol: "XRP", available: 74.62 }
        ],
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
            openQuantity: 74.62,
            entryQuantity: 74.62,
            entryPriceAvg: 0.67,
            currentPrice: 0.67,
            costBasisValue: 50,
            positionValue: 50,
            realizedPnL: 0,
            unrealizedPnL: 0,
            openedAt: "2026-03-29T12:01:00Z"
          }
        ],
        allocations: [{ marketSymbol: "XRP/USDT", value: 50, weight: 1 }]
      },
      orders: [
        {
          id: "order-1",
          walletID: "wallet-1",
          marketSymbol: "XRP/USDT",
          orderSource: "manual",
          side: "buy",
          baseQuantity: 74.62,
          quoteAmount: 50,
          expectedPrice: 0.67,
          status: "filled",
          realizedPnL: 0,
          positionAfter: 74.62,
          createdAt: "2026-03-29T12:00:00Z"
        }
      ],
      activity: [
        {
          id: "log-1",
          walletID: "wallet-1",
          marketSymbol: "XRP/USDT",
          logType: "trade",
          title: "Demo buy recorded",
          message: "Bought 74.6200 XRP at 0.6700 USDT. Position size is now updating in the portfolio view.",
          createdAt: "2026-03-29T12:01:00Z"
        }
      ]
    },
    {
      portfolio: {
        walletID: "wallet-1",
        baseCurrency: "USDT",
        accountingMode: "average_cost",
        totalValue: 10035,
        cashBalance: 9915,
        positionValue: 46.96,
        realizedPnL: 5,
        unrealizedPnL: 5.87,
        balances: [
          { walletID: "wallet-1", assetSymbol: "USDT", available: 9915 },
          { walletID: "wallet-1", assetSymbol: "XRP", available: 58.7 }
        ],
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
            openQuantity: 58.7,
            entryQuantity: 58.7,
            entryPriceAvg: 0.7,
            currentPrice: 0.8,
            costBasisValue: 41.09,
            positionValue: 46.96,
            realizedPnL: 5,
            unrealizedPnL: 5.87,
            openedAt: "2026-03-29T12:05:00Z"
          }
        ],
        allocations: [{ marketSymbol: "XRP/USDT", value: 46.96, weight: 1 }]
      },
      orders: [
        {
          id: "order-3",
          walletID: "wallet-1",
          marketSymbol: "XRP/USDT",
          orderSource: "strategy",
          strategyID: "strategy-1",
          side: "sell",
          baseQuantity: 50,
          quoteAmount: 40,
          expectedPrice: 0.8,
          status: "filled",
          realizedPnL: 5,
          positionAfter: 58.7,
          createdAt: "2026-03-29T12:04:00Z"
        }
      ],
      activity: [
        {
          id: "log-3",
          walletID: "wallet-1",
          marketSymbol: "XRP/USDT",
          logType: "trade",
          title: "Demo sell recorded",
          message: "Sold 50.0000 XRP at 0.8000 USDT. Review realized PnL in the updated portfolio and order history.",
          createdAt: "2026-03-29T12:05:00Z"
        }
      ]
    }
  ];

  return states[index];
}

async function installRoutes(page) {
  let orderStateIndex = 0;
  let strategyState = {
    id: "strategy-1",
    walletID: "wallet-1",
    marketSymbol: "XRP/USDT",
    status: "active",
    config: {
      dipBuy: { enabled: true, dipPercent: 5, spendQuoteAmount: 100 },
      takeProfit: { enabled: true, triggerPercent: 8 },
      stopLoss: { enabled: true, triggerPercent: 3 }
    },
    referencePrice: 0.64,
    lastDecision: "buy",
    lastOutcome: "executed",
    lastReason: "Dip-buy fired because XRP/USDT dropped below the configured threshold."
  };

  await page.addInitScript(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
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
            source: "fresh",
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
      const payload = request.postDataJSON() ?? {};
      if (payload.side === "sell") {
        orderStateIndex = 1;
        strategyState = {
          ...strategyState,
          lastDecision: "sell",
          lastOutcome: "executed",
          lastReason: "Take-profit fired because XRP/USDT moved above the configured target."
        };
      }

      await route.fulfill({
        status: 201,
        contentType: "application/json",
        body: JSON.stringify({ order: { id: "order-doc" } })
      });
      return;
    }

    if (url.includes("/api/v1/portfolios/")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ portfolio: state(orderStateIndex).portfolio })
      });
      return;
    }

    if (url.includes("/api/v1/orders")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ orders: state(orderStateIndex).orders })
      });
      return;
    }

    if (url.includes("/api/v1/activity")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ activity: state(orderStateIndex).activity })
      });
      return;
    }

    if (url.includes("/api/v1/strategies")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ strategies: [strategyState] })
      });
      return;
    }

    await route.fallback();
  });
}

async function capture() {
  await fs.mkdir(docsScreenshotDir, { recursive: true });

  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1600, height: 2200 } });

  await installRoutes(page);

  await page.goto(appBaseUrl);
  await page.getByRole("heading", { name: /demo execution/i }).waitFor();
  await page.screenshot({ path: path.join(docsScreenshotDir, "dashboard-overview.png"), fullPage: true });

  await page.getByRole("button", { name: "FIFO" }).click();
  await page.screenshot({ path: path.join(docsScreenshotDir, "accounting-mode-switch.png"), fullPage: true });

  await page.goto(`${appBaseUrl}/markets/XRP%2FUSDT`);
  await page.getByText(/focused trading screen/i).waitFor();
  await page.screenshot({ path: path.join(docsScreenshotDir, "market-detail-page.png"), fullPage: true });

  await page.screenshot({ path: path.join(docsScreenshotDir, "strategy-automation-active.png"), fullPage: true });

  await page.locator("input").nth(1).fill("50");
  await page.getByRole("button", { name: /run demo sell/i }).click();
  await page.getByText(/demo sell executed/i).waitFor();
  await page.screenshot({ path: path.join(docsScreenshotDir, "demo-sell-success.png"), fullPage: true });

  await browser.close();
}

capture().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
