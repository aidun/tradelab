import React from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { Hero } from "@/components/hero";

describe("Hero", () => {
  const fetchCounts: Record<string, number> = {};

  beforeEach(() => {
    for (const key of Object.keys(fetchCounts)) {
      delete fetchCounts[key];
    }

    vi.spyOn(global, "fetch").mockImplementation((input) => {
      const url = String(input);

      function record(key: string) {
        fetchCounts[key] = (fetchCounts[key] ?? 0) + 1;
      }

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
        record(`candles:${url.includes("interval=15m") ? "15m" : "1h"}`);
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
                source: url.includes("interval=15m") ? "stale" : "fresh",
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

      if (url.includes("/api/v1/orders")) {
        record("orders");
        return Promise.resolve(
          new Response(
            JSON.stringify({
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
              ]
            })
          )
        );
      }

      if (url.includes("/api/v1/activity")) {
        record("activity");
        return Promise.resolve(
          new Response(
            JSON.stringify({
              activity: [
                {
                  id: "log-1",
                  walletID: "wallet-1",
                  logType: "trade",
                  title: "Demo buy recorded",
                  message: "A demo market buy was created for XRP/USDT.",
                  createdAt: "2026-03-29T12:01:00Z"
                }
              ]
            })
          )
        );
      }

      record("portfolio");
      return Promise.resolve(
        new Response(
          JSON.stringify({
            portfolio: {
              walletID: "wallet-1",
              baseCurrency: "USDT",
              totalValue: 10000,
              cashBalance: 9950,
              balances: [
                {
                  walletID: "wallet-1",
                  assetSymbol: "USDT",
                  available: 9950
                }
              ],
              positions: []
            }
          })
        )
      );
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders the market dashboard with API-backed content", async () => {
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
  });

  it("refreshes only chart data when the interval changes", async () => {
    render(<Hero />);

    await waitFor(() => {
      expect(screen.getByText(/feed fresh/i)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: "15m" }));

    await waitFor(() => {
      expect(screen.getByText(/feed stale fallback/i)).toBeInTheDocument();
    });

    expect(fetchCounts["markets"]).toBe(1);
    expect(fetchCounts["portfolio"]).toBe(1);
    expect(fetchCounts["orders"]).toBe(1);
    expect(fetchCounts["activity"]).toBe(1);
    expect(fetchCounts["candles:1h"]).toBe(1);
    expect(fetchCounts["candles:15m"]).toBe(1);
  });
});
