"use client";

import React, { startTransition, useEffect, useState } from "react";
import { fetchMarkets, fetchPortfolio, placeMarketBuy, type Market, type PortfolioSummary } from "@/lib/api";

const DEMO_USER_ID = "cfbf7c8f-eaf9-47fa-8674-2a29fed1fcc9";
const DEMO_WALLET_ID = "1ddb1c1c-827f-4bf0-b85a-3d5786c3b26c";

function formatCurrency(value: number) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 2
  }).format(value);
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("en-US", {
    maximumFractionDigits: 4
  }).format(value);
}

export function MarketDashboard() {
  const [markets, setMarkets] = useState<Market[]>([]);
  const [portfolio, setPortfolio] = useState<PortfolioSummary | null>(null);
  const [selectedMarket, setSelectedMarket] = useState("XRP/USDT");
  const [quoteAmount, setQuoteAmount] = useState("50");
  const [expectedPrice, setExpectedPrice] = useState("0.67");
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function loadData() {
    setError(null);

    const [marketList, portfolioSummary] = await Promise.all([
      fetchMarkets(),
      fetchPortfolio(DEMO_WALLET_ID)
    ]);

    setMarkets(marketList);
    setPortfolio(portfolioSummary);
  }

  useEffect(() => {
    let cancelled = false;

    startTransition(() => {
      loadData()
        .catch((loadError: Error) => {
          if (!cancelled) {
            setError(loadError.message);
          }
        })
        .finally(() => {
          if (!cancelled) {
            setIsLoading(false);
          }
        });
    });

    return () => {
      cancelled = true;
    };
  }, []);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSuccess(null);
    setIsSubmitting(true);

    try {
      await placeMarketBuy({
        userID: DEMO_USER_ID,
        walletID: DEMO_WALLET_ID,
        marketSymbol: selectedMarket,
        quoteAmount: Number(quoteAmount),
        expectedPrice: Number(expectedPrice)
      });

      await loadData();
      setSuccess(`Demo buy executed for ${selectedMarket}.`);
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "Order failed");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="grid-glow min-h-screen overflow-hidden px-6 py-8 md:px-10 lg:px-14">
      <section className="mx-auto flex w-full max-w-7xl flex-col gap-8">
        <div className="rounded-[40px] border border-[var(--line)] bg-[var(--surface-strong)] p-6 shadow-[0_30px_90px_rgba(0,0,0,0.35)] backdrop-blur md:p-8">
          <div className="flex flex-col gap-8 lg:flex-row lg:items-start lg:justify-between">
            <div className="max-w-3xl">
              <p className="font-[var(--font-mono)] text-sm uppercase tracking-[0.32em] text-[var(--accent)]">
                TradeLab Live Sandbox
              </p>
              <h1 className="mt-4 text-5xl font-semibold leading-none tracking-[-0.05em] md:text-7xl">
                Demo execution with
                <span className="block text-[var(--accent-warm)]">real wallet movement.</span>
              </h1>
              <p className="mt-6 max-w-2xl text-lg leading-8 text-[var(--muted)]">
                Markets and portfolio state are now loaded from the Go API. Every demo buy updates the
                wallet and the open position model behind the scenes.
              </p>
            </div>

            <div className="grid min-w-[300px] gap-4 rounded-[28px] border border-[var(--line)] bg-[rgba(7,17,31,0.78)] p-5 font-[var(--font-mono)] text-sm text-[var(--muted)]">
              <div className="flex items-center justify-between gap-8">
                <span>Demo wallet</span>
                <span>{DEMO_WALLET_ID.slice(0, 8)}...</span>
              </div>
              <div className="flex items-center justify-between gap-8">
                <span>Total value</span>
                <span className="text-[var(--accent)]">
                  {portfolio ? formatCurrency(portfolio.totalValue) : "--"}
                </span>
              </div>
              <div className="flex items-center justify-between gap-8">
                <span>Cash balance</span>
                <span>{portfolio ? formatCurrency(portfolio.cashBalance) : "--"}</span>
              </div>
            </div>
          </div>

          {error ? (
            <div className="mt-6 rounded-2xl border border-[rgba(255,107,120,0.35)] bg-[rgba(255,107,120,0.08)] px-4 py-3 text-sm text-[var(--accent-hot)]">
              {error}
            </div>
          ) : null}

          {success ? (
            <div className="mt-6 rounded-2xl border border-[rgba(110,242,211,0.35)] bg-[rgba(110,242,211,0.08)] px-4 py-3 text-sm text-[var(--accent)]">
              {success}
            </div>
          ) : null}

          <div className="mt-10 grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">
                    Markets
                  </p>
                  <h2 className="mt-3 text-2xl font-semibold">Reference pairs</h2>
                </div>
                <span className="text-sm text-[var(--muted)]">
                  {isLoading ? "Loading..." : `${markets.length} tracked`}
                </span>
              </div>

              <div className="mt-6 grid gap-3">
                {markets.map((market) => (
                  <button
                    key={market.id}
                    type="button"
                    onClick={() => setSelectedMarket(market.symbol)}
                    className={`rounded-[24px] border px-4 py-4 text-left transition ${
                      selectedMarket === market.symbol
                        ? "border-[var(--accent)] bg-[rgba(110,242,211,0.08)]"
                        : "border-[var(--line)] bg-[rgba(7,17,31,0.45)]"
                    }`}
                  >
                    <div className="flex items-center justify-between gap-4">
                      <div>
                        <p className="text-lg font-semibold">{market.symbol}</p>
                        <p className="mt-1 text-sm text-[var(--muted)]">
                          {market.baseAsset} priced in {market.quoteAsset}
                        </p>
                      </div>
                      <div className="text-right text-sm text-[var(--muted)]">
                        <p>{market.exchange}</p>
                        <p className="mt-1">Min {formatCurrency(market.minNotional)}</p>
                      </div>
                    </div>
                  </button>
                ))}
              </div>
            </section>

            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">
                Demo buy ticket
              </p>
              <h2 className="mt-3 text-2xl font-semibold">Execute market buy</h2>

              <form className="mt-6 grid gap-4" onSubmit={handleSubmit}>
                <label className="grid gap-2 text-sm text-[var(--muted)]">
                  Market
                  <select
                    value={selectedMarket}
                    onChange={(event) => setSelectedMarket(event.target.value)}
                    className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.6)] px-4 py-3 text-[var(--text)] outline-none"
                  >
                    {markets.map((market) => (
                      <option key={market.id} value={market.symbol}>
                        {market.symbol}
                      </option>
                    ))}
                  </select>
                </label>

                <label className="grid gap-2 text-sm text-[var(--muted)]">
                  Quote amount (USDT)
                  <input
                    value={quoteAmount}
                    onChange={(event) => setQuoteAmount(event.target.value)}
                    className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.6)] px-4 py-3 text-[var(--text)] outline-none"
                  />
                </label>

                <label className="grid gap-2 text-sm text-[var(--muted)]">
                  Expected execution price
                  <input
                    value={expectedPrice}
                    onChange={(event) => setExpectedPrice(event.target.value)}
                    className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.6)] px-4 py-3 text-[var(--text)] outline-none"
                  />
                </label>

                <button
                  type="submit"
                  disabled={isSubmitting || isLoading}
                  className="mt-2 rounded-2xl bg-[var(--accent)] px-5 py-4 font-semibold text-[#04111a] transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {isSubmitting ? "Executing..." : "Run demo buy"}
                </button>
              </form>
            </section>
          </div>

          <div className="mt-6 grid gap-6 lg:grid-cols-2">
            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">
                Balances
              </p>
              <div className="mt-5 grid gap-3">
                {portfolio?.balances.map((balance) => (
                  <div
                    key={balance.assetSymbol}
                    className="flex items-center justify-between rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4"
                  >
                    <span className="font-medium">{balance.assetSymbol}</span>
                    <span className="font-[var(--font-mono)] text-[var(--muted)]">
                      {formatNumber(balance.available)}
                    </span>
                  </div>
                ))}
              </div>
            </section>

            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">
                Open positions
              </p>
              <div className="mt-5 grid gap-3">
                {portfolio?.positions.length ? (
                  portfolio.positions.map((position) => (
                    <div
                      key={position.id}
                      className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4"
                    >
                      <div className="flex items-center justify-between gap-4">
                        <span className="text-lg font-semibold">{position.marketSymbol}</span>
                        <span className="text-sm text-[var(--accent)]">{formatCurrency(position.positionValue)}</span>
                      </div>
                      <p className="mt-2 text-sm text-[var(--muted)]">
                        Qty {formatNumber(position.entryQuantity)} at avg {formatCurrency(position.entryPriceAvg)}
                      </p>
                    </div>
                  ))
                ) : (
                  <div className="rounded-2xl border border-dashed border-[var(--line)] px-4 py-6 text-sm text-[var(--muted)]">
                    No open positions yet. Execute the first XRP or BTC demo buy to populate this view.
                  </div>
                )}
              </div>
            </section>
          </div>
        </div>
      </section>
    </main>
  );
}
