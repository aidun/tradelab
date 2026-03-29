"use client";

import React, { startTransition, useEffect, useMemo, useState } from "react";
import {
  fetchCandles,
  placeMarketBuy,
  type Candle,
  type CandleFeed,
  type MarketDataMeta
} from "@/lib/api";
import { AuthEntryActions, AuthStatusControls, useTradeLabAuth } from "@/lib/tradelab-auth";
import { useAccountSession } from "@/lib/use-account-session";

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

function formatShortTime(value: string) {
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    hour: "numeric"
  }).format(new Date(value));
}

function intervalLabel(interval: string) {
  switch (interval) {
    case "15m":
      return "15m";
    case "1h":
      return "1h";
    case "4h":
      return "4h";
    default:
      return interval;
  }
}

function getLastItem<T>(items: T[]) {
  if (items.length === 0) {
    return null;
  }

  return items[items.length - 1];
}

function formatFeedTime(value: string) {
  return new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    minute: "2-digit",
    second: "2-digit"
  }).format(new Date(value));
}

function CandleChart({ candles }: { candles: Candle[] }) {
  if (candles.length === 0) {
    return (
      <div className="rounded-2xl border border-dashed border-[var(--line)] px-4 py-12 text-sm text-[var(--muted)]">
        Candle data will appear as soon as the market feed responds.
      </div>
    );
  }

  const width = 960;
  const height = 320;
  const paddingX = 18;
  const paddingY = 18;
  const chartWidth = width - paddingX * 2;
  const chartHeight = height - paddingY * 2;
  const candleSlot = chartWidth / candles.length;
  const candleWidth = Math.max(4, candleSlot * 0.56);

  const low = Math.min(...candles.map((item) => item.lowPrice));
  const high = Math.max(...candles.map((item) => item.highPrice));
  const priceRange = Math.max(high - low, 0.0001);

  function mapPrice(price: number) {
    const normalized = (price - low) / priceRange;
    return paddingY + chartHeight - normalized * chartHeight;
  }

  return (
    <svg
      viewBox={`0 0 ${width} ${height}`}
      role="img"
      aria-label="Live market chart"
      className="h-[320px] w-full rounded-[28px] border border-[var(--line)] bg-[linear-gradient(180deg,rgba(8,24,38,0.96),rgba(4,11,22,0.96))] p-3"
    >
      <defs>
        <linearGradient id="chartGlow" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="rgba(110,242,211,0.24)" />
          <stop offset="100%" stopColor="rgba(255,176,107,0.05)" />
        </linearGradient>
      </defs>

      <rect x="0" y="0" width={width} height={height} fill="url(#chartGlow)" rx="24" />

      {[0.2, 0.4, 0.6, 0.8].map((line) => (
        <line
          key={line}
          x1={paddingX}
          x2={width - paddingX}
          y1={paddingY + chartHeight * line}
          y2={paddingY + chartHeight * line}
          stroke="rgba(187,199,212,0.12)"
          strokeDasharray="4 8"
        />
      ))}

      {candles.map((candle, index) => {
        const x = paddingX + index * candleSlot + candleSlot / 2;
        const wickTop = mapPrice(candle.highPrice);
        const wickBottom = mapPrice(candle.lowPrice);
        const openY = mapPrice(candle.openPrice);
        const closeY = mapPrice(candle.closePrice);
        const bodyTop = Math.min(openY, closeY);
        const bodyHeight = Math.max(Math.abs(openY - closeY), 3);
        const isBull = candle.closePrice >= candle.openPrice;

        return (
          <g key={`${candle.openTime}-${index}`}>
            <line
              x1={x}
              x2={x}
              y1={wickTop}
              y2={wickBottom}
              stroke={isBull ? "rgba(110,242,211,0.92)" : "rgba(255,107,120,0.92)"}
              strokeWidth="2"
              strokeLinecap="round"
            />
            <rect
              x={x - candleWidth / 2}
              y={bodyTop}
              width={candleWidth}
              height={bodyHeight}
              rx="4"
              fill={isBull ? "rgba(110,242,211,0.95)" : "rgba(255,107,120,0.95)"}
            />
          </g>
        );
      })}

      {candles
        .filter((_, index) => index % Math.max(1, Math.floor(candles.length / 5)) === 0)
        .map((candle, index) => (
          <text
            key={`${candle.openTime}-${index}-label`}
            x={paddingX + candles.findIndex((item) => item.openTime === candle.openTime) * candleSlot + candleSlot / 2}
            y={height - 8}
            textAnchor="middle"
            fill="rgba(187,199,212,0.72)"
            fontSize="12"
          >
            {formatShortTime(candle.openTime)}
          </text>
        ))}
    </svg>
  );
}

export function MarketDashboard() {
  const auth = useTradeLabAuth();
  const {
    guestSession,
    registeredAccount,
    markets,
    portfolio,
    orders,
    activity,
    isLoading,
    isUpgrading,
    showUpgradePrompt,
    error,
    success,
    activeWalletID,
    accountModeLabel,
    shouldShowAuthValuePrompt,
    clearMessages,
    setErrorMessage,
    setSuccessMessage,
    refreshCoreData,
    upgradeGuestSession,
    activeAccessToken
  } = useAccountSession();
  const [candles, setCandles] = useState<Candle[]>([]);
  const [chartMeta, setChartMeta] = useState<MarketDataMeta | null>(null);
  const [selectedMarket, setSelectedMarket] = useState("XRP/USDT");
  const [selectedInterval, setSelectedInterval] = useState("1h");
  const [quoteAmount, setQuoteAmount] = useState("50");
  const [isChartLoading, setIsChartLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [chartError, setChartError] = useState<string | null>(null);
  const accountSummary = useMemo(() => {
    if (registeredAccount && auth.user) {
      return auth.user.displayName;
    }

    if (guestSession) {
      return `${guestSession.walletID.slice(0, 8)}...`;
    }

    return "--";
  }, [auth.user, guestSession, registeredAccount]);

  async function loadChartData() {
    setChartError(null);
    setIsChartLoading(true);

    const feed: CandleFeed = await fetchCandles(selectedMarket, selectedInterval, 48);
    setCandles(feed.candles);
    setChartMeta(feed.meta);
    setIsChartLoading(false);
  }

  useEffect(() => {
    let cancelled = false;

    if (!activeWalletID) {
      return;
    }

    // Chart refreshes stay isolated so interval or pair switches do not blank portfolio and history panels.
    startTransition(() => {
      loadChartData().catch((loadError: Error) => {
        if (!cancelled) {
          setChartError(loadError.message);
          setIsChartLoading(false);
        }
      });
    });

    return () => {
      cancelled = true;
    };
  }, [activeWalletID, selectedMarket, selectedInterval]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    clearMessages();
    setIsSubmitting(true);

    try {
      const token = await activeAccessToken();
      if (!activeWalletID) {
        throw new Error("TradeLab session is not ready yet");
      }

      await placeMarketBuy({
        token,
        marketSymbol: selectedMarket,
        quoteAmount: Number(quoteAmount)
      });

      // Successful orders need a full data refresh because balances, positions, orders, and chart-linked price hints all change together.
      await Promise.all([refreshCoreData(), loadChartData()]);
      setSuccessMessage(`Demo buy executed for ${selectedMarket}.`);
    } catch (submitError) {
      setErrorMessage(submitError instanceof Error ? submitError.message : "Order failed");
    } finally {
      setIsSubmitting(false);
    }
  }

  const latestCandle = getLastItem(candles);

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
                Live market candles now stream into the dashboard from the backend. You can inspect the
                active pair before sending a demo order and keep the chart, wallet, and trade log on one
                screen.
              </p>
            </div>

            <div className="grid min-w-[300px] gap-4 rounded-[28px] border border-[var(--line)] bg-[rgba(7,17,31,0.78)] p-5 font-[var(--font-mono)] text-sm text-[var(--muted)]">
              <div className="flex items-center justify-between gap-8">
                <span>Account mode</span>
                <span className={registeredAccount ? "text-[var(--accent)]" : "text-[var(--accent-warm)]"}>{accountModeLabel}</span>
              </div>
              <div className="flex items-center justify-between gap-8">
                <span>Account owner</span>
                <span>{accountSummary}</span>
              </div>
              <div className="flex items-center justify-between gap-8">
                <span>Total value</span>
                <span className="text-[var(--accent)]">
                  {portfolio ? formatCurrency(portfolio.totalValue) : "--"}
                </span>
              </div>
              <div className="flex items-center justify-between gap-8">
                <span>Last price</span>
                <span>{latestCandle ? formatCurrency(latestCandle.closePrice) : "--"}</span>
              </div>
              <div className="flex items-center justify-between gap-8">
                <span>Feed state</span>
                <span className={chartMeta?.source === "stale" ? "text-[var(--accent-hot)]" : "text-[var(--accent)]"}>
                  {chartMeta?.source === "stale" ? "Stale" : "Fresh"}
                </span>
              </div>
              <div className="pt-2">
                <AuthStatusControls />
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

          {shouldShowAuthValuePrompt ? (
            <section className="mt-6 rounded-[28px] border border-[var(--line)] bg-[rgba(7,17,31,0.6)] p-5">
              <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                <div>
                  <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">
                    Durable access
                  </p>
                  <h2 className="mt-2 text-2xl font-semibold">Keep this sandbox beyond the guest session.</h2>
                  <p className="mt-3 max-w-2xl text-sm leading-7 text-[var(--muted)]">
                    Sign up after you have seen the product value. Registered accounts restore the same
                    demo account across sessions and devices while the product stays demo-only.
                  </p>
                </div>
                <AuthEntryActions />
              </div>
            </section>
          ) : null}

          {showUpgradePrompt ? (
            <section className="mt-6 rounded-[28px] border border-[rgba(110,242,211,0.28)] bg-[rgba(8,24,38,0.82)] p-5">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">
                Guest upgrade
              </p>
              <h2 className="mt-2 text-2xl font-semibold">Keep your guest demo data or start fresh?</h2>
              <p className="mt-3 max-w-2xl text-sm leading-7 text-[var(--muted)]">
                You just signed in. Choose whether the current guest wallet, orders, and activity should
                become your registered demo account, or whether TradeLab should create a clean account for
                you.
              </p>
              <div className="mt-5 flex flex-wrap gap-3">
                <button
                  type="button"
                  disabled={isUpgrading}
                  onClick={() => {
                    void upgradeGuestSession(true);
                  }}
                  className="rounded-full bg-[var(--accent)] px-4 py-2 text-sm font-semibold text-[#04111a] transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {isUpgrading ? "Upgrading..." : "Keep guest demo data"}
                </button>
                <button
                  type="button"
                  disabled={isUpgrading}
                  onClick={() => {
                    void upgradeGuestSession(false);
                  }}
                  className="rounded-full border border-[var(--line)] px-4 py-2 text-sm font-medium text-[var(--text)] transition hover:border-[var(--accent)] disabled:cursor-not-allowed disabled:opacity-60"
                >
                  Start fresh
                </button>
              </div>
            </section>
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
                  Server-side execution pricing
                  <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.6)] px-4 py-3 text-[var(--text)]">
                    {latestCandle ? `${formatCurrency(latestCandle.closePrice)} estimated from live feed` : "Waiting for market data..."}
                  </div>
                </label>

                <button
                  type="submit"
                  disabled={isSubmitting || isLoading || !activeWalletID}
                  className="mt-2 rounded-2xl bg-[var(--accent)] px-5 py-4 font-semibold text-[#04111a] transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {isSubmitting ? "Executing..." : "Run demo buy"}
                </button>
              </form>
            </section>
          </div>

          <section className="mt-6 rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">
                  Live market chart
                </p>
                <h2 className="mt-3 text-2xl font-semibold">{selectedMarket}</h2>
              </div>

              <div className="flex flex-wrap items-center gap-3 text-sm text-[var(--muted)]">
                <span className="rounded-full border border-[var(--line)] px-3 py-2">
                  Last close {latestCandle ? formatCurrency(latestCandle.closePrice) : "--"}
                </span>
                {["15m", "1h", "4h"].map((interval) => (
                  <button
                    key={interval}
                    type="button"
                    onClick={() => setSelectedInterval(interval)}
                    className={`rounded-full border px-3 py-2 transition ${
                      selectedInterval === interval
                        ? "border-[var(--accent)] bg-[rgba(110,242,211,0.1)] text-[var(--accent)]"
                        : "border-[var(--line)]"
                    }`}
                  >
                    {intervalLabel(interval)}
                  </button>
                ))}
              </div>
            </div>

            <div className="mt-6">
              {isChartLoading ? (
                <div className="rounded-2xl border border-dashed border-[var(--line)] px-4 py-12 text-sm text-[var(--muted)]">
                  Loading candle data...
                </div>
              ) : chartError ? (
                <div className="rounded-2xl border border-[rgba(255,107,120,0.35)] bg-[rgba(255,107,120,0.08)] px-4 py-12 text-sm text-[var(--accent-hot)]">
                  {chartError}
                </div>
              ) : (
                <CandleChart candles={candles} />
              )}
            </div>

            <div className="mt-4 flex flex-wrap items-center gap-3 text-sm text-[var(--muted)]">
              <span className="rounded-full border border-[var(--line)] px-3 py-2">
                Feed {chartMeta?.source === "stale" ? "stale fallback" : "fresh"}
              </span>
              {chartMeta ? (
                <span className="rounded-full border border-[var(--line)] px-3 py-2">
                  Updated {formatFeedTime(chartMeta.generatedAt)}
                </span>
              ) : null}
            </div>

            <div className="mt-5 grid gap-3 md:grid-cols-3">
              <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">
                  Session high
                </p>
                <p className="mt-3 text-2xl font-semibold">
                  {latestCandle ? formatCurrency(latestCandle.highPrice) : "--"}
                </p>
              </div>
              <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">
                  Session low
                </p>
                <p className="mt-3 text-2xl font-semibold">
                  {latestCandle ? formatCurrency(latestCandle.lowPrice) : "--"}
                </p>
              </div>
              <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">
                  Last volume
                </p>
                <p className="mt-3 text-2xl font-semibold">
                  {latestCandle ? formatNumber(latestCandle.baseVolume) : "--"}
                </p>
              </div>
            </div>
          </section>

          <div className="mt-6 grid gap-6 xl:grid-cols-2">
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
                        <span className="text-sm text-[var(--accent)]">
                          {formatCurrency(position.positionValue)}
                        </span>
                      </div>
                      <p className="mt-2 text-sm text-[var(--muted)]">
                        Qty {formatNumber(position.entryQuantity)} at avg{" "}
                        {formatCurrency(position.entryPriceAvg)}
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

            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">
                Recent orders
              </p>
              <div className="mt-5 grid gap-3">
                {orders.length ? (
                  orders.map((order) => (
                    <div
                      key={order.id}
                      className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4"
                    >
                      <div className="flex items-center justify-between gap-4">
                        <span className="text-lg font-semibold">{order.marketSymbol}</span>
                        <span className="text-sm uppercase text-[var(--accent)]">{order.status}</span>
                      </div>
                      <p className="mt-2 text-sm text-[var(--muted)]">
                        Quote {formatCurrency(order.quoteAmount)} at {formatCurrency(order.expectedPrice)}
                      </p>
                    </div>
                  ))
                ) : (
                  <div className="rounded-2xl border border-dashed border-[var(--line)] px-4 py-6 text-sm text-[var(--muted)]">
                    No orders yet. The next demo buy will appear here immediately.
                  </div>
                )}
              </div>
            </section>

            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">
                Activity log
              </p>
              <div className="mt-5 grid gap-3">
                {activity.length ? (
                  activity.map((item) => (
                    <div
                      key={item.id}
                      className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4"
                    >
                      <div className="flex items-center justify-between gap-4">
                        <span className="text-lg font-semibold">{item.title}</span>
                        <span className="text-sm uppercase text-[var(--accent-hot)]">{item.logType}</span>
                      </div>
                      <p className="mt-2 text-sm text-[var(--muted)]">{item.message}</p>
                    </div>
                  ))
                ) : (
                  <div className="rounded-2xl border border-dashed border-[var(--line)] px-4 py-6 text-sm text-[var(--muted)]">
                    Activity will populate as soon as demo orders and bot actions start flowing.
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
