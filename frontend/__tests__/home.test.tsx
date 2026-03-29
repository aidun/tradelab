import React from "react";
import { render, screen, waitFor } from "@testing-library/react";
import { Hero } from "@/components/hero";

describe("Hero", () => {
  beforeEach(() => {
    vi.spyOn(global, "fetch").mockImplementation((input) => {
      const url = String(input);

      if (url.includes("/api/v1/markets")) {
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

    expect(screen.getByText(/run demo buy/i)).toBeInTheDocument();
    expect(screen.getByText(/demo buy recorded/i)).toBeInTheDocument();
  });
});
