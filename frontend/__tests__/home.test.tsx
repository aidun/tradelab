import React from "react";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { Hero } from "@/components/hero";
import { CHART_AUTO_REFRESH_MS, MarketDashboard } from "@/components/market-dashboard";

type FetchScenario = {
  candleMode?: "fresh" | "stale";
  failCandles?: boolean;
  failCandlesAfterFirst?: boolean;
  failOrder?: boolean;
  failActivityOnRefresh?: boolean;
  initialOrderStateIndex?: number;
};

function installFetchMock(scenario: FetchScenario = {}) {
  const fetchCounts: Record<string, number> = {};

  function record(key: string) {
    fetchCounts[key] = (fetchCounts[key] ?? 0) + 1;
  }

  const orderResponses = [
    {
      orders: [
        {
          id: "order-1",
          walletID: "wallet-1",
          marketSymbol: "XRP/USDT",
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
          message: "A demo market buy was created for XRP/USDT.",
          createdAt: "2026-03-29T12:01:00Z"
        }
      ],
      portfolio: {
        walletID: "wallet-1",
        baseCurrency: "USDT",
        accountingMode: "average_cost",
        totalValue: 10000,
        cashBalance: 9950,
        positionValue: 0,
        realizedPnL: 0,
        unrealizedPnL: 0,
        balances: [
          {
            walletID: "wallet-1",
            assetSymbol: "USDT",
            available: 9950
          }
        ],
        positions: [],
        allocations: []
      }
    },
    {
      orders: [
        {
          id: "order-2",
          walletID: "wallet-1",
          marketSymbol: "XRP/USDT",
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
          message: "A second demo market buy was created for XRP/USDT.",
          createdAt: "2026-03-29T12:03:00Z"
        }
      ],
      portfolio: {
        walletID: "wallet-1",
        baseCurrency: "USDT",
        accountingMode: "average_cost",
        totalValue: 10025,
        cashBalance: 9875,
        positionValue: 75,
        realizedPnL: 0,
        unrealizedPnL: 0,
        balances: [
          {
            walletID: "wallet-1",
            assetSymbol: "USDT",
            available: 9875
          }
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
            positionValue: 75,
            realizedPnL: 0,
            unrealizedPnL: 0,
            openedAt: "2026-03-29T12:02:00Z"
          }
        ],
        allocations: [{ marketSymbol: "XRP/USDT", value: 75, weight: 1 }]
      }
    }
  ];

  let orderStateIndex = scenario.initialOrderStateIndex ?? 0;

  vi.spyOn(global, "fetch").mockImplementation((input, init) => {
    const url = String(input);

    if (url.includes("/api/v1/sessions/demo")) {
      record("session");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            session: {
              id: "session-1",
              user_id: "user-1",
              wallet_id: "wallet-1",
              token: "token-1",
              expires_at: "2026-04-29T12:00:00Z"
            }
          })
        )
      );
    }

    if (url.includes("/candles")) {
      const interval = url.includes("interval=15m") ? "15m" : "1h";
      record(`candles:${interval}`);

      if (scenario.failCandles || (scenario.failCandlesAfterFirst && fetchCounts[`candles:${interval}`] > 1)) {
        return Promise.resolve(new Response(JSON.stringify({ error: "failed" }), { status: 502 }));
      }

      return Promise.resolve(
        new Response(
          JSON.stringify({
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
              source: interval === "15m" || scenario.candleMode === "stale" ? "stale" : "fresh",
              generated_at: "2026-03-29T12:05:00Z"
            }
          })
        )
      );
    }

    if (url.includes("/api/v1/markets")) {
      record("markets");
      return Promise.resolve(
        new Response(
          JSON.stringify({
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
        )
      );
    }

    if (url.includes("/api/v1/orders") && init?.method === "POST") {
      record("order-submit");

      if (scenario.failOrder) {
        return Promise.resolve(new Response(JSON.stringify({ error: "Order failed" }), { status: 422 }));
      }

      orderStateIndex = 1;
      return Promise.resolve(new Response(JSON.stringify({ order: { id: "order-2" } }), { status: 201 }));
    }

    if (url.includes("/api/v1/orders")) {
      record("orders");
      return Promise.resolve(new Response(JSON.stringify({ orders: orderResponses[orderStateIndex].orders })));
    }

    if (url.includes("/api/v1/activity")) {
      record("activity");

      if (scenario.failActivityOnRefresh && orderStateIndex > 0) {
        return Promise.resolve(new Response(JSON.stringify({ error: "Failed to load activity" }), { status: 500 }));
      }

      return Promise.resolve(new Response(JSON.stringify({ activity: orderResponses[orderStateIndex].activity })));
    }

    if (url.includes("/api/v1/strategies")) {
      record("strategies");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            strategies: []
          })
        )
      );
    }

    record("portfolio");
    return Promise.resolve(new Response(JSON.stringify({ portfolio: orderResponses[orderStateIndex].portfolio })));
  });

  return fetchCounts;
}

describe("Hero", () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
    vi.useRealTimers();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders the market dashboard with API-backed content", async () => {
    installFetchMock();

    render(<Hero />);

    expect(screen.getByRole("heading", { name: /demo execution with/i })).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getAllByText(/xrp\/usdt/i)[0]).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.getByLabelText(/live market chart/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/run demo buy/i)).toBeInTheDocument();
    expect(screen.getByText(/demo buy recorded/i)).toBeInTheDocument();
    expect(screen.getByText(/^local-dev$/i)).toBeInTheDocument();
    expect(screen.getByText(/^workspace$/i)).toBeInTheDocument();
    expect(screen.getByText(/^uncommitted$/i)).toBeInTheDocument();
  });

  it("stores guest session state in sessionStorage instead of localStorage", async () => {
    installFetchMock();

    render(<Hero />);

    await waitFor(() => {
      expect(window.sessionStorage.getItem("tradelab.demo-session")).toBeTruthy();
    });

    expect(window.localStorage.getItem("tradelab.demo-session")).toBeNull();
  });

  it("refreshes only chart data when the interval changes", async () => {
    const fetchCounts = installFetchMock();

    render(<MarketDashboard />);

    await waitFor(() => {
      expect(fetchCounts["candles:1h"]).toBe(1);
    });

    fireEvent.click(screen.getByRole("button", { name: "15m" }));

    await waitFor(() => {
      expect(fetchCounts["candles:15m"]).toBe(1);
    });

    expect(screen.getByText(/stale fallback/i)).toBeInTheDocument();

    expect(fetchCounts.markets).toBe(1);
    expect(fetchCounts.portfolio).toBe(1);
    expect(fetchCounts.orders).toBe(1);
    expect(fetchCounts.activity).toBe(1);
    expect(fetchCounts["candles:1h"]).toBe(1);
    expect(fetchCounts["candles:15m"]).toBe(1);
  });

  it("shows a success message and refreshes portfolio panels after a demo buy", async () => {
    const fetchCounts = installFetchMock();

    render(<Hero />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /run demo buy/i })).toBeEnabled();
    });

    fireEvent.click(screen.getByRole("button", { name: /run demo buy/i }));

    await waitFor(() => {
      expect(screen.getByText(/demo buy executed for xrp\/usdt/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/^demo buy executed$/i)).toBeInTheDocument();
    expect(screen.getByText(/\$10,025.00/)).toBeInTheDocument();
    expect(fetchCounts["order-submit"]).toBe(1);
    expect(fetchCounts.portfolio).toBe(2);
    expect(fetchCounts.orders).toBe(2);
    expect(fetchCounts.activity).toBe(2);
  });

  it("keeps balances visible when activity refresh fails after a demo buy", async () => {
    installFetchMock({ failActivityOnRefresh: true });

    render(<Hero />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /run demo buy/i })).toBeEnabled();
    });

    fireEvent.click(screen.getByRole("button", { name: /run demo buy/i }));

    await waitFor(() => {
      expect(screen.getByText(/demo buy executed for xrp\/usdt/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/\$10,025.00/)).toBeInTheDocument();
    expect(screen.getByText(/failed to load activity/i)).toBeInTheDocument();
    expect(screen.getByText(/\$9,875.00/)).toBeInTheDocument();
  });

  it("keeps non-chart panels visible when candle loading fails", async () => {
    installFetchMock({ failCandles: true });

    render(<Hero />);

    await waitFor(() => {
      expect(screen.getByText(/^failed$/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/demo buy recorded/i)).toBeInTheDocument();
    expect(screen.getByText(/balances/i)).toBeInTheDocument();
  });

  it("surfaces stale feed metadata from the backend", async () => {
    installFetchMock({ candleMode: "stale" });

    render(<Hero />);

    await waitFor(() => {
      expect(screen.getByText(/feed stale fallback/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/feed state/i)).toBeInTheDocument();
    expect(screen.getByText(/^stale$/i)).toBeInTheDocument();
  });

  it("auto-refreshes the market-detail chart and preserves the last candles on refresh failure", async () => {
    const fetchCounts = installFetchMock({ failCandlesAfterFirst: true });
    const setIntervalSpy = vi.spyOn(window, "setInterval");

    render(<MarketDashboard detailOnly initialMarket="XRP/USDT" />);

    await screen.findByLabelText(/live market chart/i);

    expect(fetchCounts["candles:1h"]).toBe(1);
    expect(screen.getByRole("button", { name: /refresh chart/i })).toBeInTheDocument();
    expect(setIntervalSpy).toHaveBeenCalledWith(expect.any(Function), CHART_AUTO_REFRESH_MS);

    const refreshCallback = setIntervalSpy.mock.calls.at(-1)?.[0];
    expect(typeof refreshCallback).toBe("function");

    await act(async () => {
      (refreshCallback as TimerHandler)();
      await Promise.resolve();
    });

    await waitFor(() => {
      expect(fetchCounts["candles:1h"]).toBe(2);
    });

    expect(screen.getByLabelText(/live market chart/i)).toBeInTheDocument();
    expect(screen.getByText(/^failed$/i)).toBeInTheDocument();
  });

  it("keeps the max-position sell control disabled until the current position is loaded", async () => {
    installFetchMock({ initialOrderStateIndex: 1 });

    render(<MarketDashboard detailOnly initialMarket="XRP/USDT" />);

    const maxButton = screen.getByRole("button", { name: /max position/i });
    expect(maxButton).toBeDisabled();

    await waitFor(() => {
      expect(screen.getByText(/open qty 108.7/i)).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(maxButton).toBeEnabled();
    });

    fireEvent.click(maxButton);

    expect(screen.getByLabelText(/sell quantity/i)).toHaveValue("108.7");
  });
});
