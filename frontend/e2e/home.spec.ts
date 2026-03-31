import { expect, test, type Page } from "@playwright/test";

function portfolioState(index: number) {
  const states = [
    {
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
      ],
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
      }
    },
    {
      orders: [
        {
          id: "order-2",
          walletID: "wallet-1",
          marketSymbol: "XRP/USDT",
          orderSource: "strategy",
          strategyID: "strategy-1",
          side: "buy",
          baseQuantity: 108.7,
          quoteAmount: 75,
          expectedPrice: 0.69,
          status: "filled",
          realizedPnL: 0,
          positionAfter: 108.7,
          createdAt: "2026-03-29T12:02:00Z"
        }
      ],
      activity: [
        {
          id: "log-2",
          walletID: "wallet-1",
          marketSymbol: "XRP/USDT",
          logType: "trade",
          title: "Demo buy executed",
          message: "Strategy executed a buy for 108.7000 XRP at 0.6900 USDT.",
          createdAt: "2026-03-29T12:03:00Z"
        }
      ],
      portfolio: {
        walletID: "wallet-1",
        baseCurrency: "USDT",
        accountingMode: "average_cost",
        totalValue: 10025,
        cashBalance: 9875,
        positionValue: 150,
        realizedPnL: 0,
        unrealizedPnL: 0,
        balances: [
          { walletID: "wallet-1", assetSymbol: "USDT", available: 9875 },
          { walletID: "wallet-1", assetSymbol: "XRP", available: 108.7 }
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
            openQuantity: 108.7,
            entryQuantity: 108.7,
            entryPriceAvg: 0.69,
            currentPrice: 0.69,
            costBasisValue: 75,
            positionValue: 150,
            realizedPnL: 0,
            unrealizedPnL: 75,
            openedAt: "2026-03-29T12:03:00Z"
          }
        ],
        allocations: [{ marketSymbol: "XRP/USDT", value: 150, weight: 1 }]
      }
    },
    {
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
          message: "Strategy executed a sell for 50.0000 XRP at 0.8000 USDT.",
          createdAt: "2026-03-29T12:05:00Z"
        }
      ],
      portfolio: {
        walletID: "wallet-1",
        baseCurrency: "USDT",
        accountingMode: "average_cost",
        totalValue: 10035,
        cashBalance: 9915,
        positionValue: 120,
        realizedPnL: 5,
        unrealizedPnL: 10,
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
      }
    },
    {
      orders: [
        {
          id: "order-4",
          walletID: "wallet-1",
          marketSymbol: "XRP/USDT",
          orderSource: "strategy",
          strategyID: "strategy-1",
          side: "sell",
          baseQuantity: 58.7,
          quoteAmount: 52.83,
          expectedPrice: 0.9,
          status: "filled",
          realizedPnL: 12,
          positionAfter: 0,
          createdAt: "2026-03-29T12:06:00Z"
        }
      ],
      activity: [
        {
          id: "log-4",
          walletID: "wallet-1",
          marketSymbol: "XRP/USDT",
          logType: "trade",
          title: "Demo sell recorded",
          message: "Strategy executed a sell for 58.7000 XRP at 0.9000 USDT.",
          createdAt: "2026-03-29T12:07:00Z"
        }
      ],
      portfolio: {
        walletID: "wallet-1",
        baseCurrency: "USDT",
        accountingMode: "average_cost",
        totalValue: 10060,
        cashBalance: 9967.83,
        positionValue: 0,
        realizedPnL: 17,
        unrealizedPnL: 0,
        balances: [
          { walletID: "wallet-1", assetSymbol: "USDT", available: 9967.83 },
          { walletID: "wallet-1", assetSymbol: "XRP", available: 0 }
        ],
        positions: [],
        allocations: []
      }
    }
  ];

  return states[index];
}

test.beforeEach(async ({ page }) => {
  let orderStateIndex = 0;
  let strategyState = {
    id: "strategy-1",
    walletID: "wallet-1",
    marketSymbol: "XRP/USDT",
    status: "draft",
    config: {
      dipBuy: { enabled: true, dipPercent: 5, spendQuoteAmount: 100 },
      takeProfit: { enabled: true, triggerPercent: 8 },
      stopLoss: { enabled: true, triggerPercent: 3 }
    },
    referencePrice: 0.64,
    lastDecision: "",
    lastOutcome: "",
    lastReason: ""
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
            source: url.includes("interval=15m") ? "stale" : "fresh",
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
      const payload = request.postDataJSON() as { side?: "buy" | "sell"; base_quantity?: number };
      if (payload.side === "buy") {
        orderStateIndex = 1;
      } else if ((payload.base_quantity ?? 0) >= 58.7) {
        orderStateIndex = 3;
      } else {
        orderStateIndex = 2;
      }

      await route.fulfill({
        status: 201,
        contentType: "application/json",
        body: JSON.stringify({ order: { id: `order-${orderStateIndex + 1}` } })
      });
      return;
    }

    if (url.endsWith("/api/v1/strategies") && request.method() === "POST") {
      const payload = request.postDataJSON() as { status: "draft" | "active" | "paused"; config: typeof strategyState.config };
      strategyState = {
        ...strategyState,
        status: payload.status,
        config: payload.config
      };

      if (payload.status === "active") {
        orderStateIndex = 1;
        strategyState = {
          ...strategyState,
          status: "active",
          lastDecision: "buy",
          lastOutcome: "executed",
          lastReason: "Dip-buy fired because XRP/USDT dropped below the configured threshold."
        };
      }

      await route.fulfill({
        status: 201,
        contentType: "application/json",
        body: JSON.stringify({ strategy: strategyState })
      });
      return;
    }

    if (url.includes("/api/v1/strategies/") && request.method() === "PATCH") {
      const payload = request.postDataJSON() as { status: "active" | "paused"; config: typeof strategyState.config };
      strategyState = {
        ...strategyState,
        status: payload.status,
        config: payload.config
      };

      if (payload.status === "active") {
        if (orderStateIndex <= 1) {
          orderStateIndex = 2;
          strategyState = {
            ...strategyState,
            lastDecision: "sell",
            lastOutcome: "executed",
            lastReason: "Take-profit fired because XRP/USDT moved above the configured target."
          };
        }
      } else if (payload.status === "paused") {
        strategyState = {
          ...strategyState,
          lastDecision: "hold",
          lastOutcome: "skipped",
          lastReason: "Strategy paused by user."
        };
      }

      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ strategy: strategyState })
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

    if (url.includes("/api/v1/backtests")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          backtest: {
            marketSymbol: "XRP/USDT",
            baseAsset: "XRP",
            quoteAsset: "USDT",
            interval: "1h",
            startTime: "2026-03-10T00:00:00Z",
            endTime: "2026-03-20T23:59:59Z",
            initialCash: 10000,
            finalCash: 10080,
            finalPositionQty: 0,
            finalPositionValue: 0,
            finalEquity: 10080,
            orders: [
              {
                id: "backtest-order-1",
                walletID: "wallet-1",
                marketSymbol: "XRP/USDT",
                orderSource: "strategy",
                side: "buy",
                baseQuantity: 150,
                quoteAmount: 100,
                expectedPrice: 0.6667,
                status: "filled",
                realizedPnL: 0,
                positionAfter: 150,
                createdAt: "2026-03-11T12:00:00Z"
              },
              {
                id: "backtest-order-2",
                walletID: "wallet-1",
                marketSymbol: "XRP/USDT",
                orderSource: "strategy",
                side: "sell",
                baseQuantity: 150,
                quoteAmount: 180,
                expectedPrice: 1.2,
                status: "filled",
                realizedPnL: 80,
                positionAfter: 0,
                createdAt: "2026-03-12T12:00:00Z"
              }
            ],
            equityCurve: [
              {
                time: "2026-03-11T12:00:00Z",
                price: 0.6667,
                cashBalance: 9900,
                openQuantity: 150,
                positionValue: 100,
                totalEquity: 10000,
                drawdownPercent: 0
              },
              {
                time: "2026-03-12T12:00:00Z",
                price: 1.2,
                cashBalance: 10080,
                openQuantity: 0,
                positionValue: 0,
                totalEquity: 10080,
                drawdownPercent: 0
              }
            ],
            summary: {
              returnPercent: 0.8,
              tradeCount: 2,
              sellCount: 1,
              winningTradeCount: 1,
              hitRatePercent: 100,
              maxDrawdownPercent: 0
            }
          }
        })
      });
      return;
    }

    if (url.includes("/api/v1/portfolios/")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ portfolio: portfolioState(orderStateIndex).portfolio })
      });
      return;
    }

    if (url.includes("/api/v1/orders")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ orders: portfolioState(orderStateIndex).orders })
      });
      return;
    }

    if (url.includes("/api/v1/activity")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ activity: portfolioState(orderStateIndex).activity })
      });
      return;
    }

    if (url.endsWith("/api/v1/account/logout")) {
      await route.fulfill({
        status: 204,
        headers: {
          "Set-Cookie": "tradelab_app_session=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax"
        },
        body: ""
      });
      return;
    }

    await route.fallback();
  });
});

async function continueAsGuest(page: Page) {
  await page.getByRole("button", { name: /continue as guest/i }).click();
  await expect(page.getByRole("heading", { name: /financial boundary/i })).toBeVisible();
  await page.getByRole("button", { name: /acknowledge and continue/i }).click();
}

test("creates a demo session and renders the dashboard", async ({ page }) => {
  await page.goto("/");

  await expect(page.getByRole("heading", { name: /access your trading workspace/i })).toBeVisible();
  await continueAsGuest(page);
  await expect(page.getByText("XRP/USDT").first()).toBeVisible();
  await expect(page.getByRole("button", { name: /run demo buy/i })).toBeVisible();
});

test("opens the dedicated market detail page", async ({ page }) => {
  await page.goto("/markets/XRP%2FUSDT");
  await expect(page.getByRole("heading", { name: /access your trading workspace/i })).toBeVisible();
  await continueAsGuest(page);

  await expect(page).toHaveURL(/\/markets\/XRP%2FUSDT|\/markets\/XRP\/USDT/);
  await expect(page.getByText(/focused trading screen/i)).toBeVisible();
  await expect(page.getByRole("button", { name: /run demo sell/i })).toBeVisible();
});

test("refreshes chart data from the market-detail header", async ({ page }) => {
  let candleRequests = 0;

  await page.route("**/api/v1/markets/*/candles**", async (route) => {
    candleRequests += 1;
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
  });

  await page.goto("/markets/XRP%2FUSDT");
  await continueAsGuest(page);
  await expect(page.getByRole("button", { name: /refresh chart/i })).toBeVisible();
  await expect.poll(() => candleRequests).toBe(1);

  await page.getByRole("button", { name: /refresh chart/i }).click();

  await expect.poll(() => candleRequests).toBe(2);
});

test("executes a demo buy and refreshes portfolio metrics", async ({ page }) => {
  await page.goto("/");
  await continueAsGuest(page);
  await page.getByRole("button", { name: /run demo buy/i }).click();

  await expect(page.getByText(/demo buy executed for xrp\/usdt/i)).toBeVisible();
  await expect(page.getByText(/\$10,025.00/)).toBeVisible();
});

test("keeps balances visible when the activity refresh fails after a demo buy", async ({ page }) => {
  let activityRequests = 0;

  await page.route("**/api/v1/activity", async (route) => {
    activityRequests += 1;
    if (activityRequests > 1) {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "Failed to load activity" })
      });
      return;
    }

    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ activity: portfolioState(0).activity })
    });
  });

  await page.goto("/");
  await continueAsGuest(page);
  await page.getByRole("button", { name: /run demo buy/i }).click();

  await expect(page.getByText(/demo buy executed for xrp\/usdt/i)).toBeVisible();
  await expect(page.getByText(/\$10,025.00/)).toBeVisible();
  await expect(page.getByText(/failed to load activity/i)).toBeVisible();
});

test("executes a partial sell from the market detail page", async ({ page }) => {
  await page.goto("/markets/XRP%2FUSDT");
  await continueAsGuest(page);
  await page.getByLabel(/sell quantity/i).fill("50");
  await page.getByRole("button", { name: /run demo sell/i }).click();

  await expect(page.getByText(/demo sell executed for xrp\/usdt/i)).toBeVisible();
  await expect(page.getByText(/open qty 58.7/i)).toBeVisible();
  await expect(page.getByText(/realized \$5.00/i).first()).toBeVisible();
});

test("closes the position with a max sell", async ({ page }) => {
  await page.goto("/markets/XRP%2FUSDT");
  await continueAsGuest(page);
  await expect(page.getByText(/open qty 74.62/i)).toBeVisible();
  await expect(page.getByRole("button", { name: /max position/i })).toBeEnabled();
  await page.getByRole("button", { name: /max position/i }).click();
  await expect(page.getByLabel(/sell quantity/i)).toHaveValue("74.62");
  await page.getByRole("button", { name: /run demo sell/i }).click();

  await expect(page.getByText(/demo sell executed for xrp\/usdt/i)).toBeVisible();
  await expect(page.getByText(/no open positions yet/i)).toBeVisible();
});

test("switches accounting modes globally", async ({ page }) => {
  await page.goto("/");
  await continueAsGuest(page);
  await page.getByRole("button", { name: "FIFO" }).click();

  await expect(page.getByText(/fifo/i).first()).toBeVisible();
});

test("configures and activates a dip-buy strategy", async ({ page }) => {
  await page.goto("/markets/XRP%2FUSDT");
  await continueAsGuest(page);
  await page.getByRole("button", { name: /activate/i }).click();

  await expect(page.getByText(/automation activated for xrp\/usdt/i)).toBeVisible();
  await expect(page.getByText(/automated/i).first()).toBeVisible();
  await expect(page.getByText(/^active$/i)).toBeVisible();
});

test("executes a strategy-driven take-profit sell after activation", async ({ page }) => {
  await page.goto("/markets/XRP%2FUSDT");
  await continueAsGuest(page);
  await page.getByRole("button", { name: /activate/i }).click();
  await page.getByRole("button", { name: /activate/i }).click();

  await expect(page.getByText(/take-profit fired/i)).toBeVisible();
  await expect(page.getByText(/realized \$5.00/i).first()).toBeVisible();
});

test("pauses an active strategy without executing a new trade", async ({ page }) => {
  await page.goto("/markets/XRP%2FUSDT");
  await continueAsGuest(page);
  await page.getByRole("button", { name: /activate/i }).click();
  await page.getByRole("button", { name: /pause/i }).click();

  await expect(page.getByText(/strategy paused by user/i)).toBeVisible();
  await expect(page.getByText(/^paused$/i)).toBeVisible();
});

test("runs a read-only backtest from the market detail page", async ({ page }) => {
  await page.goto("/markets/XRP%2FUSDT");
  await continueAsGuest(page);
  await page.getByRole("button", { name: /run backtest/i }).click();

  await expect(page.getByText(/backtest ready for xrp\/usdt/i)).toBeVisible();
  await expect(page.getByText(/strategy sell/i)).toBeVisible();
  await expect(page.getByText(/max drawdown/i)).toBeVisible();
  await expect(page.getByText(/recent backtests/i)).toBeVisible();
  await expect(page.getByRole("img", { name: /backtest equity curve/i })).toBeVisible();
});
