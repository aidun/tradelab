"use client";

import Link from "next/link";
import React, { startTransition, useEffect, useMemo, useState } from "react";

import { fetchCandles, placeMarketOrder, type AccountingMode, type Candle, type MarketDataMeta } from "@/lib/api";
import { AuthEntryActions, AuthStatusControls, useTradeLabAuth } from "@/lib/tradelab-auth";
import { useAccountSession } from "@/lib/use-account-session";

type MarketDashboardProps = {
  detailOnly?: boolean;
  initialMarket?: string;
};

function formatCurrency(value: number) {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 2 }).format(value);
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("en-US", { maximumFractionDigits: 4 }).format(value);
}

function formatFeedTime(value: string) {
  return new Intl.DateTimeFormat("en-US", { hour: "numeric", minute: "2-digit", second: "2-digit" }).format(new Date(value));
}

function accountingModeLabel(mode: AccountingMode) {
  switch (mode) {
    case "fifo":
      return "FIFO";
    case "hybrid":
      return "Hybrid";
    default:
      return "Average cost";
  }
}

function pnlTone(value: number) {
  if (value > 0) return "text-[var(--accent)]";
  if (value < 0) return "text-[var(--accent-hot)]";
  return "text-[var(--muted)]";
}

function CandleChart({ candles }: { candles: Candle[] }) {
  if (candles.length === 0) {
    return <div className="rounded-2xl border border-dashed border-[var(--line)] px-4 py-12 text-sm text-[var(--muted)]">Candle data will appear as soon as the market feed responds.</div>;
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
  const range = Math.max(high - low, 0.0001);
  const mapPrice = (price: number) => paddingY + chartHeight - ((price - low) / range) * chartHeight;

  return (
    <svg viewBox={`0 0 ${width} ${height}`} role="img" aria-label="Live market chart" className="h-[320px] w-full rounded-[28px] border border-[var(--line)] bg-[linear-gradient(180deg,rgba(8,24,38,0.96),rgba(4,11,22,0.96))] p-3">
      <rect x="0" y="0" width={width} height={height} fill="rgba(8,24,38,0.96)" rx="24" />
      {candles.map((candle, index) => {
        const x = paddingX + index * candleSlot + candleSlot / 2;
        const isBull = candle.closePrice >= candle.openPrice;
        return (
          <g key={`${candle.openTime}-${index}`}>
            <line x1={x} x2={x} y1={mapPrice(candle.highPrice)} y2={mapPrice(candle.lowPrice)} stroke={isBull ? "rgba(110,242,211,0.92)" : "rgba(255,107,120,0.92)"} strokeWidth="2" />
            <rect x={x - candleWidth / 2} y={Math.min(mapPrice(candle.openPrice), mapPrice(candle.closePrice))} width={candleWidth} height={Math.max(Math.abs(mapPrice(candle.openPrice) - mapPrice(candle.closePrice)), 3)} rx="4" fill={isBull ? "rgba(110,242,211,0.95)" : "rgba(255,107,120,0.95)"} />
          </g>
        );
      })}
    </svg>
  );
}

export function MarketDashboard({ detailOnly = false, initialMarket = "XRP/USDT" }: MarketDashboardProps) {
  const auth = useTradeLabAuth();
  const {
    guestSession, registeredAccount, markets, portfolio, orders, activity, accountingMode, isLoading, isUpgrading,
    showUpgradePrompt, error, success, activeWalletID, accountModeLabel, shouldShowAuthValuePrompt,
    clearMessages, setErrorMessage, setSuccessMessage, setAccountingMode, refreshCoreData, upgradeGuestSession, activeAccessToken
  } = useAccountSession();
  const [selectedMarket, setSelectedMarket] = useState(initialMarket);
  const [selectedInterval, setSelectedInterval] = useState("1h");
  const [buyQuoteAmount, setBuyQuoteAmount] = useState("50");
  const [sellBaseQuantity, setSellBaseQuantity] = useState("25");
  const [candles, setCandles] = useState<Candle[]>([]);
  const [chartMeta, setChartMeta] = useState<MarketDataMeta | null>(null);
  const [chartError, setChartError] = useState<string | null>(null);
  const [isChartLoading, setIsChartLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    setSelectedMarket(initialMarket);
  }, [initialMarket]);

  useEffect(() => {
    if (!activeWalletID) return;
    let cancelled = false;
    startTransition(() => {
      fetchCandles(selectedMarket, selectedInterval, 48)
        .then((feed) => {
          if (!cancelled) {
            setCandles(feed.candles);
            setChartMeta(feed.meta);
            setChartError(null);
            setIsChartLoading(false);
          }
        })
        .catch((loadError: Error) => {
          if (!cancelled) {
            setChartError(loadError.message);
            setIsChartLoading(false);
          }
        });
    });
    return () => {
      cancelled = true;
    };
  }, [activeWalletID, selectedInterval, selectedMarket]);

  const selectedPosition = useMemo(() => portfolio?.positions.find((position) => position.marketSymbol === selectedMarket) ?? null, [portfolio, selectedMarket]);
  const visibleOrders = useMemo(() => (detailOnly ? orders.filter((order) => order.marketSymbol === selectedMarket) : orders), [detailOnly, orders, selectedMarket]);
  const visibleActivity = useMemo(() => (detailOnly ? activity.filter((item) => item.marketSymbol === selectedMarket || item.marketSymbol === "") : activity), [activity, detailOnly, selectedMarket]);
  const accountSummary = registeredAccount && auth.user ? auth.user.displayName : guestSession ? `${guestSession.walletID.slice(0, 8)}...` : "--";
  const lastPrice = candles.length > 0 ? candles[candles.length - 1].closePrice : null;

  async function submitOrder(side: "buy" | "sell") {
    clearMessages();
    setIsSubmitting(true);

    try {
      const token = await activeAccessToken();
      await placeMarketOrder({
        side,
        marketSymbol: selectedMarket,
        quoteAmount: side === "buy" ? Number(buyQuoteAmount) : undefined,
        baseQuantity: side === "sell" ? Number(sellBaseQuantity) : undefined,
        token
      });
      await refreshCoreData();
      setSuccessMessage(`${side === "buy" ? "Demo buy" : "Demo sell"} executed for ${selectedMarket}.`);
    } catch (submitError) {
      setErrorMessage(submitError instanceof Error ? submitError.message : "Order failed");
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
              {detailOnly ? <Link href="/" className="font-[var(--font-mono)] text-sm uppercase tracking-[0.32em] text-[var(--accent)]">Back to overview</Link> : <p className="font-[var(--font-mono)] text-sm uppercase tracking-[0.32em] text-[var(--accent)]">TradeLab Live Sandbox</p>}
              <h1 className="mt-4 text-5xl font-semibold leading-none tracking-[-0.05em] md:text-7xl">{detailOnly ? selectedMarket : "Demo execution with real wallet movement."}</h1>
              <p className="mt-6 max-w-2xl text-lg leading-8 text-[var(--muted)]">{detailOnly ? "Focused trading screen with chart, current position, buy and sell controls, and filtered history." : "Portfolio overview, accounting-mode selection, and fast navigation into market detail."}</p>
            </div>
            <div className="grid min-w-[320px] gap-4 rounded-[28px] border border-[var(--line)] bg-[rgba(7,17,31,0.78)] p-5 font-[var(--font-mono)] text-sm text-[var(--muted)]">
              <div className="flex items-center justify-between gap-8"><span>Account mode</span><span className={registeredAccount ? "text-[var(--accent)]" : "text-[var(--accent-warm)]"}>{accountModeLabel}</span></div>
              <div className="flex items-center justify-between gap-8"><span>Account owner</span><span>{accountSummary}</span></div>
              <div className="flex items-center justify-between gap-8"><span>Accounting mode</span><span className="text-[var(--accent)]">{accountingModeLabel(accountingMode)}</span></div>
              <div className="flex items-center justify-between gap-8"><span>Total value</span><span className="text-[var(--accent)]">{portfolio ? formatCurrency(portfolio.totalValue) : "--"}</span></div>
              <div className="flex items-center justify-between gap-8"><span>Feed state</span><span className={chartMeta?.source === "stale" ? "text-[var(--accent-hot)]" : "text-[var(--accent)]"}>{chartMeta?.source === "stale" ? "Stale" : "Fresh"}</span></div>
              <div className="pt-2"><AuthStatusControls /></div>
            </div>
          </div>

          <div className="mt-6 flex flex-wrap gap-3">
            {(["average_cost", "fifo", "hybrid"] as AccountingMode[]).map((mode) => (
              <button key={mode} type="button" onClick={() => setAccountingMode(mode)} className={`rounded-full border px-3 py-2 text-sm transition ${accountingMode === mode ? "border-[var(--accent)] bg-[rgba(110,242,211,0.1)] text-[var(--accent)]" : "border-[var(--line)] text-[var(--muted)]"}`}>
                {accountingModeLabel(mode)}
              </button>
            ))}
          </div>

          {error ? <div className="mt-6 rounded-2xl border border-[rgba(255,107,120,0.35)] bg-[rgba(255,107,120,0.08)] px-4 py-3 text-sm text-[var(--accent-hot)]">{error}</div> : null}
          {success ? <div className="mt-6 rounded-2xl border border-[rgba(110,242,211,0.35)] bg-[rgba(110,242,211,0.08)] px-4 py-3 text-sm text-[var(--accent)]">{success}</div> : null}

          {shouldShowAuthValuePrompt ? (
            <section className="mt-6 rounded-[28px] border border-[var(--line)] bg-[rgba(7,17,31,0.6)] p-5">
              <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                <div><h2 className="text-2xl font-semibold">Keep this sandbox beyond the guest session.</h2><p className="mt-3 max-w-2xl text-sm leading-7 text-[var(--muted)]">Registered accounts restore the same demo account across sessions and devices.</p></div>
                <AuthEntryActions />
              </div>
            </section>
          ) : null}

          {showUpgradePrompt ? (
            <section className="mt-6 rounded-[28px] border border-[rgba(110,242,211,0.28)] bg-[rgba(8,24,38,0.82)] p-5">
              <h2 className="text-2xl font-semibold">Keep your guest demo data or start fresh?</h2>
              <div className="mt-5 flex flex-wrap gap-3">
                <button type="button" disabled={isUpgrading} onClick={() => void upgradeGuestSession(true)} className="rounded-full bg-[var(--accent)] px-4 py-2 text-sm font-semibold text-[#04111a] disabled:opacity-60">{isUpgrading ? "Upgrading..." : "Keep guest demo data"}</button>
                <button type="button" disabled={isUpgrading} onClick={() => void upgradeGuestSession(false)} className="rounded-full border border-[var(--line)] px-4 py-2 text-sm font-medium text-[var(--text)] disabled:opacity-60">Start fresh</button>
              </div>
            </section>
          ) : null}

          <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-5">
            <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Cash balance</p><p className="mt-3 text-2xl font-semibold">{portfolio ? formatCurrency(portfolio.cashBalance ?? 0) : "--"}</p></div>
            <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Position value</p><p className="mt-3 text-2xl font-semibold">{portfolio ? formatCurrency(portfolio.positionValue ?? 0) : "--"}</p></div>
            <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Realized PnL</p><p className={`mt-3 text-2xl font-semibold ${portfolio ? pnlTone(portfolio.realizedPnL ?? 0) : ""}`}>{portfolio ? formatCurrency(portfolio.realizedPnL ?? 0) : "--"}</p></div>
            <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Unrealized PnL</p><p className={`mt-3 text-2xl font-semibold ${portfolio ? pnlTone(portfolio.unrealizedPnL ?? 0) : ""}`}>{portfolio ? formatCurrency(portfolio.unrealizedPnL ?? 0) : "--"}</p></div>
            <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Last price</p><p className="mt-3 text-2xl font-semibold">{lastPrice ? formatCurrency(lastPrice) : "--"}</p></div>
          </div>

          {!detailOnly ? (
            <section className="mt-6 rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <div className="grid gap-3">
                {markets.map((market) => (
                  <Link key={market.id} href={`/markets/${encodeURIComponent(market.symbol)}`} className="rounded-[24px] border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4 text-left transition hover:border-[var(--accent)]">
                    <div className="flex items-center justify-between gap-4">
                      <div><p className="text-lg font-semibold">{market.symbol}</p><p className="mt-1 text-sm text-[var(--muted)]">{market.baseAsset} priced in {market.quoteAsset}</p></div>
                      <span className="text-sm text-[var(--accent)]">Open market detail</span>
                    </div>
                  </Link>
                ))}
              </div>
              <form className="mt-6 grid gap-3 rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4" onSubmit={(event) => { event.preventDefault(); void submitOrder("buy"); }}>
                <p className="text-lg font-semibold">Quick buy</p>
                <input value={buyQuoteAmount} onChange={(event) => setBuyQuoteAmount(event.target.value)} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.6)] px-4 py-3 text-[var(--text)] outline-none" />
                <button type="submit" disabled={isSubmitting || !activeWalletID} className="rounded-2xl bg-[var(--accent)] px-5 py-4 font-semibold text-[#04111a] disabled:opacity-60">{isSubmitting ? "Executing..." : "Run demo buy"}</button>
              </form>
            </section>
          ) : null}

          <section className="mt-6 rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
            <div className="flex items-center justify-between gap-4">
              <h2 className="text-2xl font-semibold">{selectedMarket}</h2>
              {!detailOnly ? (
                <select value={selectedMarket} onChange={(event) => setSelectedMarket(event.target.value)} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.6)] px-4 py-3 text-[var(--text)] outline-none">
                  {markets.map((market) => <option key={market.id} value={market.symbol}>{market.symbol}</option>)}
                </select>
              ) : null}
            </div>
            <div className="mt-4 flex flex-wrap gap-3">{["15m", "1h", "4h"].map((interval) => <button key={interval} type="button" onClick={() => setSelectedInterval(interval)} className={`rounded-full border px-3 py-2 text-sm ${selectedInterval === interval ? "border-[var(--accent)] bg-[rgba(110,242,211,0.1)] text-[var(--accent)]" : "border-[var(--line)]"}`}>{interval}</button>)}</div>
            <div className="mt-6">{isChartLoading ? <div className="rounded-2xl border border-dashed border-[var(--line)] px-4 py-12 text-sm text-[var(--muted)]">Loading candle data...</div> : chartError ? <div className="rounded-2xl border border-[rgba(255,107,120,0.35)] bg-[rgba(255,107,120,0.08)] px-4 py-12 text-sm text-[var(--accent-hot)]">{chartError}</div> : <CandleChart candles={candles} />}</div>
            <div className="mt-4 flex flex-wrap items-center gap-3 text-sm text-[var(--muted)]">
              <span className="rounded-full border border-[var(--line)] px-3 py-2">Feed {chartMeta?.source === "stale" ? "stale fallback" : "fresh"}</span>
              {chartMeta ? <span className="rounded-full border border-[var(--line)] px-3 py-2">Updated {formatFeedTime(chartMeta.generatedAt)}</span> : null}
            </div>
          </section>

          <div className="mt-6 grid gap-6 xl:grid-cols-2">
            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">{detailOnly ? "Trade tickets" : "Balances"}</p>
              {detailOnly ? (
                <div className="mt-5 grid gap-4 md:grid-cols-2">
                  <form className="grid gap-3 rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4" onSubmit={(event) => { event.preventDefault(); void submitOrder("buy"); }}>
                    <p className="text-lg font-semibold">Buy {selectedMarket}</p>
                    <input value={buyQuoteAmount} onChange={(event) => setBuyQuoteAmount(event.target.value)} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.6)] px-4 py-3 text-[var(--text)] outline-none" />
                    <button type="submit" disabled={isSubmitting || !activeWalletID} className="rounded-2xl bg-[var(--accent)] px-5 py-4 font-semibold text-[#04111a] disabled:opacity-60">{isSubmitting ? "Executing..." : "Run demo buy"}</button>
                  </form>
                  <form className="grid gap-3 rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4" onSubmit={(event) => { event.preventDefault(); void submitOrder("sell"); }}>
                    <p className="text-lg font-semibold">Sell {selectedMarket}</p>
                    <input value={sellBaseQuantity} onChange={(event) => setSellBaseQuantity(event.target.value)} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.6)] px-4 py-3 text-[var(--text)] outline-none" />
                    <button type="button" onClick={() => selectedPosition && setSellBaseQuantity(String(selectedPosition.openQuantity))} className="rounded-2xl border border-[var(--line)] px-4 py-3 text-sm font-medium text-[var(--muted)]">Max position</button>
                    <button type="submit" disabled={isSubmitting || !activeWalletID || !selectedPosition} className="rounded-2xl bg-[var(--accent-warm)] px-5 py-4 font-semibold text-[#04111a] disabled:opacity-60">{isSubmitting ? "Executing..." : "Run demo sell"}</button>
                  </form>
                </div>
              ) : (
                <div className="mt-5 grid gap-3">{portfolio?.balances.map((balance) => <div key={balance.assetSymbol} className="flex items-center justify-between rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4"><span className="font-medium">{balance.assetSymbol}</span><span className="font-[var(--font-mono)] text-[var(--muted)]">{formatNumber(balance.available)}</span></div>)}</div>
              )}
            </section>

            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">{detailOnly ? "Current market position" : "Open positions"}</p>
              <div className="mt-5 grid gap-3">
                {(detailOnly ? (selectedPosition ? [selectedPosition] : []) : portfolio?.positions ?? []).length ? (
                  (detailOnly ? (selectedPosition ? [selectedPosition] : []) : portfolio?.positions ?? []).map((position) => (
                    <div key={position.marketSymbol} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                      <div className="flex items-center justify-between gap-4"><span className="text-lg font-semibold">{position.marketSymbol}</span><span className="text-sm text-[var(--accent)]">{formatCurrency(position.positionValue)}</span></div>
                      <p className="mt-2 text-sm text-[var(--muted)]">Open qty {formatNumber(position.openQuantity)} at avg {formatCurrency(position.entryPriceAvg)}</p>
                      <div className="mt-3 grid gap-2 text-sm text-[var(--muted)] md:grid-cols-2">
                        <span>Cost basis {formatCurrency(position.costBasisValue)}</span>
                        <span>Current {formatCurrency(position.currentPrice)}</span>
                        <span className={pnlTone(position.realizedPnL)}>Realized {formatCurrency(position.realizedPnL)}</span>
                        <span className={pnlTone(position.unrealizedPnL)}>Unrealized {formatCurrency(position.unrealizedPnL)}</span>
                      </div>
                    </div>
                  ))
                ) : <div className="rounded-2xl border border-dashed border-[var(--line)] px-4 py-6 text-sm text-[var(--muted)]">No open positions yet.</div>}
              </div>
            </section>

            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">{detailOnly ? "Selected market order history" : "Recent orders"}</p>
              <div className="mt-5 grid gap-3">
                {visibleOrders.length ? visibleOrders.map((order) => (
                  <div key={order.id} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                    <div className="flex items-center justify-between gap-4"><span className="text-lg font-semibold">{order.marketSymbol}</span><span className={order.side === "sell" ? "text-[var(--accent-warm)]" : "text-[var(--accent)]"}>{order.side}</span></div>
                    <p className="mt-2 text-sm text-[var(--muted)]">{formatNumber(order.baseQuantity)} units at {formatCurrency(order.expectedPrice)}</p>
                    <div className="mt-3 flex flex-wrap gap-3 text-sm text-[var(--muted)]"><span>Quote {formatCurrency(order.quoteAmount)}</span><span>After trade {formatNumber(order.positionAfter)}</span>{order.side === "sell" ? <span className={pnlTone(order.realizedPnL)}>Realized {formatCurrency(order.realizedPnL)}</span> : null}</div>
                  </div>
                )) : <div className="rounded-2xl border border-dashed border-[var(--line)] px-4 py-6 text-sm text-[var(--muted)]">No orders yet for this scope.</div>}
              </div>
            </section>

            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">{detailOnly ? "Selected market activity" : "Activity log"}</p>
              <div className="mt-5 grid gap-3">
                {visibleActivity.length ? visibleActivity.map((item) => (
                  <div key={item.id} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                    <div className="flex items-center justify-between gap-4"><span className="text-lg font-semibold">{item.title}</span><span className="text-sm uppercase text-[var(--accent-hot)]">{item.logType}</span></div>
                    <p className="mt-2 text-sm text-[var(--muted)]">{item.message}</p>
                  </div>
                )) : <div className="rounded-2xl border border-dashed border-[var(--line)] px-4 py-6 text-sm text-[var(--muted)]">Activity will populate as soon as trades are executed for this scope.</div>}
              </div>
            </section>
          </div>
        </div>
      </section>
    </main>
  );
}
