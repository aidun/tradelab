"use client";

import Link from "next/link";
import React, { startTransition, useCallback, useEffect, useMemo, useRef, useState } from "react";

import { fetchCandles, patchStrategy, placeMarketOrder, runBacktest, saveStrategy, type AccountingMode, type BacktestRun, type Candle, type MarketDataMeta, type Strategy, type StrategyConfig } from "@/lib/api";
import { resolveBuildInfo } from "@/lib/build-info";
import { AuthEntryActions, AuthStatusControls, useTradeLabAuth } from "@/lib/tradelab-auth";
import { useAccountSession } from "@/lib/use-account-session";

type MarketDashboardProps = {
  detailOnly?: boolean;
  initialMarket?: string;
  autoStartGuest?: boolean;
};

type BacktestReportEntry = {
  id: string;
  label: string;
  run: BacktestRun;
};

export const CHART_AUTO_REFRESH_MS = 30_000;

const baseControlClassName =
  "focus-ring rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.6)] px-4 py-3 text-[var(--text)] hover:border-[rgba(110,242,211,0.4)]";
const secondaryButtonClassName =
  "focus-ring rounded-2xl border border-[var(--line)] px-4 py-3 text-sm font-medium text-[var(--text)] hover:border-[var(--accent)] hover:text-[var(--accent)] disabled:cursor-not-allowed disabled:opacity-60";
const accentButtonClassName =
  "focus-ring rounded-2xl bg-[var(--accent)] px-4 py-3 text-sm font-semibold text-[#04111a] hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60";
const warmButtonClassName =
  "focus-ring rounded-2xl bg-[var(--accent-warm)] px-4 py-3 text-sm font-semibold text-[#04111a] hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-60";

function EmptyPanelState({ children }: { children: React.ReactNode }) {
  return <div className="rounded-2xl border border-dashed border-[var(--line)] px-4 py-6 text-sm text-[var(--muted)]">{children}</div>;
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 2 }).format(value);
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("en-US", { maximumFractionDigits: 4 }).format(value);
}

function formatFeedTime(value: string) {
  return new Intl.DateTimeFormat("en-US", { hour: "numeric", minute: "2-digit", second: "2-digit" }).format(new Date(value));
}

function isoDate(daysOffset: number) {
  const value = new Date();
  value.setUTCDate(value.getUTCDate() + daysOffset);
  return value.toISOString().slice(0, 10);
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

function strategyStatusTone(status: Strategy["status"]) {
  switch (status) {
    case "active":
      return "text-[var(--accent)]";
    case "paused":
      return "text-[var(--accent-warm)]";
    case "archived":
      return "text-[var(--accent-hot)]";
    default:
      return "text-[var(--muted)]";
  }
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

function BacktestEquityChart({ points }: { points: BacktestRun["equityCurve"] }) {
  if (points.length === 0) {
    return <EmptyPanelState>The equity curve will appear after a completed backtest run.</EmptyPanelState>;
  }

  const width = 960;
  const height = 220;
  const padding = 18;
  const values = points.map((point) => point.totalEquity);
  const low = Math.min(...values);
  const high = Math.max(...values);
  const range = Math.max(high - low, 0.0001);
  const step = points.length > 1 ? (width - padding * 2) / (points.length - 1) : 0;
  const path = points
    .map((point, index) => {
      const x = padding + step * index;
      const y = padding + (height - padding * 2) - ((point.totalEquity - low) / range) * (height - padding * 2);
      return `${index === 0 ? "M" : "L"} ${x.toFixed(2)} ${y.toFixed(2)}`;
    })
    .join(" ");

  return (
    <svg viewBox={`0 0 ${width} ${height}`} role="img" aria-label="Backtest equity curve" className="h-[220px] w-full rounded-[28px] border border-[var(--line)] bg-[linear-gradient(180deg,rgba(8,24,38,0.96),rgba(4,11,22,0.96))] p-3">
      <rect x="0" y="0" width={width} height={height} fill="rgba(8,24,38,0.96)" rx="24" />
      <path d={path} fill="none" stroke="rgba(110,242,211,0.95)" strokeWidth="4" strokeLinecap="round" />
    </svg>
  );
}
/** MarketDashboard renders the overview and market-detail trading experience. */
export function MarketDashboard({
  detailOnly = false,
  initialMarket = "XRP/USDT",
  autoStartGuest = true
}: MarketDashboardProps) {
  const auth = useTradeLabAuth();
  const buildInfo = resolveBuildInfo();
  const {
    guestSession, registeredAccount, markets, portfolio, orders, activity, activityError, strategies, accountingMode, isLoading, isUpgrading,
    showUpgradePrompt, error, success, activeWalletID, accountModeLabel, shouldShowAuthValuePrompt,
    clearMessages, setErrorMessage, setSuccessMessage, setAccountingMode, refreshCoreData, upgradeGuestSession, activeAccessToken
  } = useAccountSession({ autoStartGuest });
  const availableStrategies = strategies ?? [];
  const [selectedMarket, setSelectedMarket] = useState(initialMarket);
  const [selectedInterval, setSelectedInterval] = useState("1h");
  const [buyQuoteAmount, setBuyQuoteAmount] = useState("50");
  const [sellBaseQuantity, setSellBaseQuantity] = useState("25");
  const sellBaseQuantityRef = React.useRef(sellBaseQuantity);
  const candleCountRef = useRef(0);
  const isChartRequestInFlight = useRef(false);
  const pendingChartReloadRef = useRef(false);
  const selectedMarketRef = useRef(selectedMarket);
  const selectedIntervalRef = useRef(selectedInterval);
  const [candles, setCandles] = useState<Candle[]>([]);
  const [chartMeta, setChartMeta] = useState<MarketDataMeta | null>(null);
  const [chartError, setChartError] = useState<string | null>(null);
  const [isChartLoading, setIsChartLoading] = useState(true);
  const [isChartRefreshing, setIsChartRefreshing] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSavingStrategy, setIsSavingStrategy] = useState(false);
  const [backtestStartDate, setBacktestStartDate] = useState(() => isoDate(-14));
  const [backtestEndDate, setBacktestEndDate] = useState(() => isoDate(-1));
  const [backtestResult, setBacktestResult] = useState<BacktestRun | null>(null);
  const [backtestHistory, setBacktestHistory] = useState<BacktestReportEntry[]>([]);
  const [backtestError, setBacktestError] = useState<string | null>(null);
  const [isRunningBacktest, setIsRunningBacktest] = useState(false);
  const [strategyConfig, setStrategyConfig] = useState<StrategyConfig>({
    dipBuy: { enabled: true, dipPercent: 5, spendQuoteAmount: 100 },
    takeProfit: { enabled: true, triggerPercent: 8 },
    stopLoss: { enabled: true, triggerPercent: 3 }
  });

  useEffect(() => {
    setSelectedMarket(initialMarket);
  }, [initialMarket]);

  useEffect(() => {
    selectedMarketRef.current = selectedMarket;
  }, [selectedMarket]);

  useEffect(() => {
    selectedIntervalRef.current = selectedInterval;
  }, [selectedInterval]);

  useEffect(() => {
    sellBaseQuantityRef.current = sellBaseQuantity;
  }, [sellBaseQuantity]);

  useEffect(() => {
    candleCountRef.current = candles.length;
  }, [candles.length]);

  const loadCandles = useCallback(async (options?: { preserveExisting?: boolean }) => {
    if (!activeWalletID) {
      return;
    }

    if (isChartRequestInFlight.current) {
      pendingChartReloadRef.current = true;
      return;
    }

    const preserveExisting = options?.preserveExisting ?? false;
    const hasRenderedCandles = candleCountRef.current > 0;
    const requestMarket = selectedMarket;
    const requestInterval = selectedInterval;
    isChartRequestInFlight.current = true;
    pendingChartReloadRef.current = false;
    setChartError(null);
    setIsChartRefreshing(true);
    if (!preserveExisting || !hasRenderedCandles) {
      setIsChartLoading(true);
    }

    try {
      const feed = await fetchCandles(selectedMarket, selectedInterval, 48);
      setCandles(feed.candles);
      setChartMeta(feed.meta);
      setChartError(null);
    } catch (loadError) {
      const nextMessage = loadError instanceof Error ? loadError.message : "We couldn't refresh the chart right now.";
      setChartError(nextMessage);
      if (!hasRenderedCandles) {
        setCandles([]);
      }
    } finally {
      isChartRequestInFlight.current = false;
      setIsChartLoading(false);
      setIsChartRefreshing(false);

      const selectionChanged = selectedMarketRef.current !== requestMarket || selectedIntervalRef.current !== requestInterval;
      if (pendingChartReloadRef.current || selectionChanged) {
        pendingChartReloadRef.current = false;
        startTransition(() => {
          void loadCandles({ preserveExisting: candleCountRef.current > 0 });
        });
      }
    }
  }, [activeWalletID, selectedInterval, selectedMarket]);

  useEffect(() => {
    if (!activeWalletID) {
      return;
    }

    startTransition(() => {
      void loadCandles({ preserveExisting: false });
    });
  }, [activeWalletID, loadCandles]);

  useEffect(() => {
    if (!detailOnly || !activeWalletID) {
      return;
    }

    const intervalHandle = window.setInterval(() => {
      startTransition(() => {
        void loadCandles({ preserveExisting: true });
      });
    }, CHART_AUTO_REFRESH_MS);

    return () => {
      window.clearInterval(intervalHandle);
    };
  }, [activeWalletID, detailOnly, loadCandles]);

  const selectedPosition = useMemo(() => portfolio?.positions.find((position) => position.marketSymbol === selectedMarket) ?? null, [portfolio, selectedMarket]);
  const selectedStrategy = useMemo(() => availableStrategies.find((item) => item.marketSymbol === selectedMarket) ?? null, [availableStrategies, selectedMarket]);
  const visibleOrders = useMemo(() => (detailOnly ? orders.filter((order) => order.marketSymbol === selectedMarket) : orders), [detailOnly, orders, selectedMarket]);
  const visibleActivity = useMemo(() => (detailOnly ? activity.filter((item) => item.marketSymbol === selectedMarket || item.marketSymbol === "") : activity), [activity, detailOnly, selectedMarket]);
  const canUseMaxSell = Boolean(selectedPosition && selectedPosition.openQuantity > 0);
  const accountSummary = registeredAccount && auth.user ? auth.user.displayName : guestSession ? `${guestSession.walletID.slice(0, 8)}…` : "--";
  const lastPrice = candles.length > 0 ? candles[candles.length - 1].closePrice : null;
  const chartRefreshLabel = isChartRefreshing ? "Refreshing…" : "Refresh Now";

  useEffect(() => {
    if (selectedStrategy) {
      setStrategyConfig(selectedStrategy.config);
      return;
    }

    setStrategyConfig({
      dipBuy: { enabled: true, dipPercent: 5, spendQuoteAmount: 100 },
      takeProfit: { enabled: true, triggerPercent: 8 },
      stopLoss: { enabled: true, triggerPercent: 3 }
    });
  }, [selectedStrategy]);

  async function submitOrder(side: "buy" | "sell") {
    clearMessages();
    setIsSubmitting(true);

    try {
      const token = await activeAccessToken();
      await placeMarketOrder({
        side,
        marketSymbol: selectedMarket,
        quoteAmount: side === "buy" ? Number(buyQuoteAmount) : undefined,
        baseQuantity: side === "sell" ? Number(sellBaseQuantityRef.current) : undefined,
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

  async function persistStrategy(nextStatus: Strategy["status"]) {
    clearMessages();
    setIsSavingStrategy(true);

    try {
      const token = await activeAccessToken();
      if (selectedStrategy) {
        await patchStrategy({
          id: selectedStrategy.id,
          status: nextStatus,
          config: strategyConfig,
          token
        });
      } else {
        await saveStrategy({
          marketSymbol: selectedMarket,
          status: nextStatus,
          config: strategyConfig,
          token
        });
      }
      await refreshCoreData();
      setSuccessMessage(`Automation ${nextStatus === "active" ? "activated" : "saved"} for ${selectedMarket}.`);
    } catch (strategyError) {
      setErrorMessage(strategyError instanceof Error ? strategyError.message : "Failed to save automation");
    } finally {
      setIsSavingStrategy(false);
    }
  }

  async function runStrategyBacktest() {
    clearMessages();
    setBacktestError(null);
    setIsRunningBacktest(true);

    try {
      const token = await activeAccessToken();
      const result = await runBacktest({
        marketSymbol: selectedMarket,
        interval: selectedInterval,
        startTime: new Date(`${backtestStartDate}T00:00:00Z`).toISOString(),
        endTime: new Date(`${backtestEndDate}T23:59:59Z`).toISOString(),
        config: strategyConfig,
        token
      });
      setBacktestResult(result);
      setBacktestHistory((current) => [
        {
          id: `${selectedMarket}-${backtestStartDate}-${backtestEndDate}-${current.length + 1}`,
          label: `${backtestStartDate} to ${backtestEndDate}`,
          run: result
        },
        ...current
      ].slice(0, 5));
      setSuccessMessage(`Backtest ready for ${selectedMarket}.`);
    } catch (runError) {
      setBacktestError(runError instanceof Error ? runError.message : "Backtest failed");
    } finally {
      setIsRunningBacktest(false);
    }
  }

  return (
    <main id="main-content" className="grid-glow min-h-screen overflow-hidden px-6 py-8 md:px-10 lg:px-14">
      <section className="mx-auto flex w-full max-w-7xl flex-col gap-8">
        <div className="rounded-[40px] border border-[var(--line)] bg-[var(--surface-strong)] p-6 shadow-[0_30px_90px_rgba(0,0,0,0.35)] backdrop-blur md:p-8">
          <div className="flex flex-col gap-8 lg:flex-row lg:items-start lg:justify-between">
            <div className="max-w-3xl">
              {detailOnly ? <Link href="/" className="focus-ring inline-flex rounded-full font-[var(--font-mono)] text-sm uppercase tracking-[0.32em] text-[var(--accent)] hover:text-[var(--text)]">Back to Overview</Link> : <p className="font-[var(--font-mono)] text-sm uppercase tracking-[0.32em] text-[var(--accent)]">TradeLab Live Sandbox</p>}
              <h1 className="mt-4 text-5xl font-semibold leading-none tracking-[-0.05em] text-balance md:text-7xl">{detailOnly ? selectedMarket : "Demo execution with real wallet movement."}</h1>
              <p className="mt-6 max-w-2xl text-lg leading-8 text-[var(--muted)]">{detailOnly ? "Focused trading screen with chart, current position, buy and sell controls, and filtered history." : "Portfolio overview, accounting-mode selection, and fast navigation into market detail."}</p>
            </div>
            <div className="grid min-w-[320px] gap-4 rounded-[28px] border border-[var(--line)] bg-[rgba(7,17,31,0.78)] p-5 font-[var(--font-mono)] text-sm text-[var(--muted)]">
              <div className="flex items-center justify-between gap-8"><span>Account mode</span><span className={registeredAccount ? "text-[var(--accent)]" : "text-[var(--accent-warm)]"}>{accountModeLabel}</span></div>
              <div className="flex items-center justify-between gap-8"><span>Account owner</span><span>{accountSummary}</span></div>
              <div className="flex items-center justify-between gap-8"><span>Accounting mode</span><span className="text-[var(--accent)]">{accountingModeLabel(accountingMode)}</span></div>
              <div className="flex items-center justify-between gap-8"><span>Total value</span><span className="text-[var(--accent)]">{portfolio ? formatCurrency(portfolio.totalValue) : "--"}</span></div>
              <div className="flex items-center justify-between gap-8"><span>Feed state</span><span className={chartMeta?.source === "stale" ? "text-[var(--accent-hot)]" : "text-[var(--accent)]"}>{chartMeta?.source === "stale" ? "Stale" : "Fresh"}</span></div>
              <div className="flex items-center justify-between gap-8"><span>Release</span><span>{buildInfo.release}</span></div>
              <div className="flex items-center justify-between gap-8"><span>Branch</span><span>{buildInfo.branch}</span></div>
              <div className="flex items-center justify-between gap-8"><span>Commit</span><span>{buildInfo.shortCommit}</span></div>
              <div className="pt-2"><AuthStatusControls /></div>
            </div>
          </div>

          <div className="mt-6 flex flex-wrap gap-3">
            {(["average_cost", "fifo", "hybrid"] as AccountingMode[]).map((mode) => (
              <button key={mode} type="button" onClick={() => setAccountingMode(mode)} className={`focus-ring rounded-full border px-3 py-2 text-sm transition-[border-color,background-color,color] ${accountingMode === mode ? "border-[var(--accent)] bg-[rgba(110,242,211,0.1)] text-[var(--accent)]" : "border-[var(--line)] text-[var(--muted)] hover:border-[var(--accent)] hover:text-[var(--text)]"}`}>
                {accountingModeLabel(mode)}
              </button>
            ))}
          </div>

          {error ? <div role="status" aria-live="polite" className="mt-6 rounded-2xl border border-[rgba(255,107,120,0.35)] bg-[rgba(255,107,120,0.08)] px-4 py-3 text-sm text-[var(--accent-hot)]">{error}</div> : null}
          {success ? <div role="status" aria-live="polite" className="mt-6 rounded-2xl border border-[rgba(110,242,211,0.35)] bg-[rgba(110,242,211,0.08)] px-4 py-3 text-sm text-[var(--accent)]">{success}</div> : null}

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
                <button type="button" disabled={isUpgrading} onClick={() => void upgradeGuestSession(true)} className="focus-ring rounded-full bg-[var(--accent)] px-4 py-2 text-sm font-semibold text-[#04111a] hover:brightness-105 disabled:opacity-60">{isUpgrading ? "Upgrading…" : "Keep guest demo data"}</button>
                <button type="button" disabled={isUpgrading} onClick={() => void upgradeGuestSession(false)} className="focus-ring rounded-full border border-[var(--line)] px-4 py-2 text-sm font-medium text-[var(--text)] hover:border-[var(--accent)] hover:text-[var(--accent)] disabled:opacity-60">Start fresh</button>
              </div>
            </section>
          ) : null}

          <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-5">
            <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Cash balance</p><p className="tabular-data mt-3 text-2xl font-semibold">{portfolio ? formatCurrency(portfolio.cashBalance ?? 0) : "--"}</p></div>
            <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Position value</p><p className="tabular-data mt-3 text-2xl font-semibold">{portfolio ? formatCurrency(portfolio.positionValue ?? 0) : "--"}</p></div>
            <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Realized PnL</p><p className={`tabular-data mt-3 text-2xl font-semibold ${portfolio ? pnlTone(portfolio.realizedPnL ?? 0) : ""}`}>{portfolio ? formatCurrency(portfolio.realizedPnL ?? 0) : "--"}</p></div>
            <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Unrealized PnL</p><p className={`tabular-data mt-3 text-2xl font-semibold ${portfolio ? pnlTone(portfolio.unrealizedPnL ?? 0) : ""}`}>{portfolio ? formatCurrency(portfolio.unrealizedPnL ?? 0) : "--"}</p></div>
            <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Last price</p><p className="tabular-data mt-3 text-2xl font-semibold">{lastPrice ? formatCurrency(lastPrice) : "--"}</p></div>
          </div>

          {!detailOnly ? (
            <section className="mt-6 rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <div className="grid gap-3">
                {markets.map((market) => (
                  <Link key={market.id} href={`/markets/${encodeURIComponent(market.symbol)}`} className="focus-ring rounded-[24px] border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4 text-left transition-[border-color,transform] hover:border-[var(--accent)] hover:-translate-y-0.5">
                    <div className="flex items-center justify-between gap-4">
                      <div className="min-w-0"><p className="text-lg font-semibold">{market.symbol}</p><p className="mt-1 text-sm text-[var(--muted)]">{market.baseAsset} priced in {market.quoteAsset}</p></div>
                      <div className="text-right">
                        <p className={`text-sm ${strategyStatusTone(availableStrategies.find((item) => item.marketSymbol === market.symbol)?.status ?? "draft")}`}>
                          {availableStrategies.find((item) => item.marketSymbol === market.symbol)?.status ?? "no automation"}
                        </p>
                        <span className="text-sm text-[var(--accent)]">Open Market Detail</span>
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
              <form className="mt-6 grid gap-3 rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4" onSubmit={(event) => { event.preventDefault(); void submitOrder("buy"); }}>
                <p className="text-lg font-semibold">Quick buy</p>
                <label className="grid gap-2 text-sm text-[var(--muted)]" htmlFor="quick-buy-quote-amount">
                  Quick buy amount (USDT)
                  <input id="quick-buy-quote-amount" name="quick-buy-quote-amount" type="number" inputMode="decimal" step="any" autoComplete="off" value={buyQuoteAmount} onChange={(event) => setBuyQuoteAmount(event.target.value)} className={`${baseControlClassName} tabular-data`} />
                </label>
                <button type="submit" disabled={isSubmitting || !activeWalletID} className="focus-ring rounded-2xl bg-[var(--accent)] px-5 py-4 font-semibold text-[#04111a] hover:brightness-105 disabled:opacity-60">{isSubmitting ? "Executing…" : "Run Demo Buy"}</button>
              </form>
            </section>
          ) : null}

          <section className="mt-6 rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
            <div className="flex items-center justify-between gap-4">
              <h2 className="text-2xl font-semibold">{selectedMarket}</h2>
              {!detailOnly ? (
                <label className="grid gap-2 text-sm text-[var(--muted)]">
                  Active market
                <select aria-label="Active market" name="active-market" value={selectedMarket} onChange={(event) => setSelectedMarket(event.target.value)} className={`${baseControlClassName} min-w-[13rem]`}>
                  {markets.map((market) => <option key={market.id} value={market.symbol}>{market.symbol}</option>)}
                </select>
                </label>
              ) : (
                <button
                  type="button"
                  aria-label="Refresh chart"
                  onClick={() => {
                    startTransition(() => {
                      void loadCandles({ preserveExisting: true });
                    });
                  }}
                  disabled={isChartRefreshing}
                  className={secondaryButtonClassName}
                >
                  {chartRefreshLabel}
                </button>
              )}
            </div>
            <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
              <div className="flex flex-wrap gap-3">{["15m", "1h", "4h"].map((interval) => <button key={interval} type="button" onClick={() => setSelectedInterval(interval)} className={`focus-ring rounded-full border px-3 py-2 text-sm transition-[border-color,background-color,color] ${selectedInterval === interval ? "border-[var(--accent)] bg-[rgba(110,242,211,0.1)] text-[var(--accent)]" : "border-[var(--line)] text-[var(--muted)] hover:border-[var(--accent)] hover:text-[var(--text)]"}`}>{interval}</button>)}</div>
              <div className="flex flex-wrap items-center gap-3 text-sm text-[var(--muted)]">
                <span className="rounded-full border border-[var(--line)] px-3 py-2">Feed {chartMeta?.source === "stale" ? "stale fallback" : "fresh"}</span>
                {chartMeta ? <span className="rounded-full border border-[var(--line)] px-3 py-2">Updated {formatFeedTime(chartMeta.generatedAt)}</span> : null}
                {detailOnly ? <span className="rounded-full border border-[var(--line)] px-3 py-2">{isChartRefreshing ? "Refreshing…" : `Auto ${Math.floor(CHART_AUTO_REFRESH_MS / 1000)}s`}</span> : null}
              </div>
            </div>
            <div className="mt-6">
              {isChartLoading && candles.length === 0 ? (
                <div className="rounded-2xl border border-dashed border-[var(--line)] px-4 py-12 text-sm text-[var(--muted)]">Loading candle data…</div>
              ) : chartError && candles.length === 0 ? (
                <div role="status" aria-live="polite" className="rounded-2xl border border-[rgba(255,107,120,0.35)] bg-[rgba(255,107,120,0.08)] px-4 py-12 text-sm text-[var(--accent-hot)]">{chartError}</div>
              ) : (
                <CandleChart candles={candles} />
              )}
            </div>
            {chartError && candles.length > 0 ? (
              <div role="status" aria-live="polite" className="mt-4 rounded-2xl border border-[rgba(255,107,120,0.35)] bg-[rgba(255,107,120,0.08)] px-4 py-4 text-sm text-[var(--accent-hot)]">
                {chartError}
              </div>
            ) : null}
          </section>

          <div className="mt-6 grid gap-6 xl:grid-cols-2">
            {detailOnly ? (
              <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur xl:col-span-2">
                <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">Automation</p>
                <div className="mt-5 grid gap-5 lg:grid-cols-[1.2fr_0.8fr]">
                  <div className="grid gap-4 md:grid-cols-3">
                    <label className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] p-4">
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-sm font-semibold">Dip buy</span>
                        <input aria-label="Enable dip-buy rule" type="checkbox" checked={strategyConfig.dipBuy.enabled} onChange={(event) => setStrategyConfig((current) => ({ ...current, dipBuy: { ...current.dipBuy, enabled: event.target.checked } }))} />
                      </div>
                      <div className="mt-3 grid gap-3">
                        <span className="grid gap-2 text-sm text-[var(--muted)]">Dip threshold (%)<input name="dip-buy-percent" type="number" inputMode="decimal" step="any" autoComplete="off" value={strategyConfig.dipBuy.dipPercent} onChange={(event) => setStrategyConfig((current) => ({ ...current, dipBuy: { ...current.dipBuy, dipPercent: Number(event.target.value) } }))} className={`${baseControlClassName} tabular-data`} /></span>
                        <span className="grid gap-2 text-sm text-[var(--muted)]">Spend amount (USDT)<input name="dip-buy-spend" type="number" inputMode="decimal" step="any" autoComplete="off" value={strategyConfig.dipBuy.spendQuoteAmount} onChange={(event) => setStrategyConfig((current) => ({ ...current, dipBuy: { ...current.dipBuy, spendQuoteAmount: Number(event.target.value) } }))} className={`${baseControlClassName} tabular-data`} /></span>
                      </div>
                    </label>
                    <label className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] p-4">
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-sm font-semibold">Take profit</span>
                        <input aria-label="Enable take-profit rule" type="checkbox" checked={strategyConfig.takeProfit.enabled} onChange={(event) => setStrategyConfig((current) => ({ ...current, takeProfit: { ...current.takeProfit, enabled: event.target.checked } }))} />
                      </div>
                      <div className="mt-3 grid gap-3">
                        <span className="grid gap-2 text-sm text-[var(--muted)]">Target gain (%)<input name="take-profit-percent" type="number" inputMode="decimal" step="any" autoComplete="off" value={strategyConfig.takeProfit.triggerPercent} onChange={(event) => setStrategyConfig((current) => ({ ...current, takeProfit: { ...current.takeProfit, triggerPercent: Number(event.target.value) } }))} className={`${baseControlClassName} tabular-data`} /></span>
                      </div>
                    </label>
                    <label className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] p-4">
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-sm font-semibold">Stop loss</span>
                        <input aria-label="Enable stop-loss rule" type="checkbox" checked={strategyConfig.stopLoss.enabled} onChange={(event) => setStrategyConfig((current) => ({ ...current, stopLoss: { ...current.stopLoss, enabled: event.target.checked } }))} />
                      </div>
                      <div className="mt-3 grid gap-3">
                        <span className="grid gap-2 text-sm text-[var(--muted)]">Stop threshold (%)<input name="stop-loss-percent" type="number" inputMode="decimal" step="any" autoComplete="off" value={strategyConfig.stopLoss.triggerPercent} onChange={(event) => setStrategyConfig((current) => ({ ...current, stopLoss: { ...current.stopLoss, triggerPercent: Number(event.target.value) } }))} className={`${baseControlClassName} tabular-data`} /></span>
                      </div>
                    </label>
                  </div>
                  <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] p-4">
                    <div className="flex items-center justify-between gap-4">
                      <span className="text-sm font-semibold">Status</span>
                      <span className={strategyStatusTone(selectedStrategy?.status ?? "draft")}>{selectedStrategy?.status ?? "draft"}</span>
                    </div>
                    <div className="mt-4 grid gap-2 text-sm text-[var(--muted)]">
                      <span>Last decision: {selectedStrategy?.lastDecision || "--"}</span>
                      <span>Last outcome: {selectedStrategy?.lastOutcome || "--"}</span>
                      <span className="tabular-data">Reference price: {selectedStrategy?.referencePrice ? formatCurrency(selectedStrategy.referencePrice) : "--"}</span>
                    </div>
                    <p className="mt-4 text-sm leading-7 text-[var(--muted)]">{selectedStrategy?.lastReason || "No evaluation yet. Save and activate automation to start the strategy loop."}</p>
                    <div className="mt-5 flex flex-wrap gap-3">
                      <button type="button" onClick={() => void persistStrategy(selectedStrategy?.status === "active" ? "active" : "draft")} disabled={isSavingStrategy} className={secondaryButtonClassName}>{isSavingStrategy ? "Saving…" : "Save bundle"}</button>
                      <button type="button" onClick={() => void persistStrategy("active")} disabled={isSavingStrategy} className={accentButtonClassName}>Activate</button>
                      <button type="button" onClick={() => void persistStrategy("paused")} disabled={isSavingStrategy || !selectedStrategy} className={warmButtonClassName}>Pause</button>
                    </div>
                  </div>
                </div>
              </section>
            ) : null}
            {detailOnly ? (
              <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur xl:col-span-2">
                <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
                  <div className="max-w-2xl">
                    <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">Backtesting</p>
                    <h3 className="mt-3 text-2xl font-semibold">Replay this strategy over a historical range.</h3>
                    <p className="mt-3 text-sm leading-7 text-[var(--muted)]">This run is read-only. It uses the current strategy settings, the selected interval, and historical candles for {selectedMarket}.</p>
                  </div>
                  <div className="grid min-w-[280px] gap-3 lg:w-[360px]">
                    <label className="grid gap-2 text-sm text-[var(--muted)]" htmlFor="backtest-start-date">
                      Start date
                      <input id="backtest-start-date" name="backtest-start-date" type="date" value={backtestStartDate} max={backtestEndDate} onChange={(event) => setBacktestStartDate(event.target.value)} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.6)] px-4 py-3 text-[var(--text)] outline-none transition focus:border-[var(--accent)] focus-visible:ring-2 focus-visible:ring-[rgba(110,242,211,0.28)]" />
                    </label>
                    <label className="grid gap-2 text-sm text-[var(--muted)]" htmlFor="backtest-end-date">
                      End date
                      <input id="backtest-end-date" name="backtest-end-date" type="date" value={backtestEndDate} min={backtestStartDate} onChange={(event) => setBacktestEndDate(event.target.value)} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.6)] px-4 py-3 text-[var(--text)] outline-none transition focus:border-[var(--accent)] focus-visible:ring-2 focus-visible:ring-[rgba(110,242,211,0.28)]" />
                    </label>
                    <button type="button" onClick={() => void runStrategyBacktest()} disabled={isRunningBacktest || !activeWalletID} className="rounded-2xl border border-[var(--line)] px-4 py-3 text-sm font-medium text-[var(--text)] transition hover:border-[rgba(110,242,211,0.32)] hover:bg-[rgba(12,36,55,0.72)] focus-visible:ring-2 focus-visible:ring-[rgba(110,242,211,0.28)] disabled:cursor-not-allowed disabled:opacity-60">
                      {isRunningBacktest ? "Running backtest…" : "Run Backtest"}
                    </button>
                  </div>
                </div>
                {backtestError ? <div role="status" aria-live="polite" className="mt-5 rounded-2xl border border-[rgba(255,107,120,0.35)] bg-[rgba(255,107,120,0.08)] px-4 py-4 text-sm text-[var(--accent-hot)]">{backtestError}</div> : null}
                {backtestResult ? (
                  <div className="mt-5 grid gap-4">
                    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                      <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Return</p><p className={`tabular-data mt-3 text-2xl font-semibold ${pnlTone(backtestResult.summary.returnPercent)}`}>{formatNumber(backtestResult.summary.returnPercent)}%</p></div>
                      <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Trades</p><p className="tabular-data mt-3 text-2xl font-semibold">{backtestResult.summary.tradeCount}</p></div>
                      <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Hit rate</p><p className="tabular-data mt-3 text-2xl font-semibold">{formatNumber(backtestResult.summary.hitRatePercent)}%</p></div>
                      <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4"><p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Max drawdown</p><p className="tabular-data mt-3 text-2xl font-semibold">{formatNumber(backtestResult.summary.maxDrawdownPercent)}%</p></div>
                    </div>
                    <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                      <div className="flex items-center justify-between gap-3">
                        <div>
                          <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Equity curve</p>
                          <p className="mt-2 text-sm text-[var(--muted)]">Track how simulated equity rose, pulled back, and closed across the selected historical window.</p>
                        </div>
                        <span className="tabular-data text-sm text-[var(--muted)]">Final equity {formatCurrency(backtestResult.finalEquity)}</span>
                      </div>
                      <div className="mt-4">
                        <BacktestEquityChart points={backtestResult.equityCurve} />
                      </div>
                    </div>
                    <div className="grid gap-4 lg:grid-cols-[0.9fr_1.1fr]">
                      <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                        <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Run snapshot</p>
                        <div className="mt-4 grid gap-2 text-sm text-[var(--muted)]">
                          <span className="tabular-data">Initial cash {formatCurrency(backtestResult.initialCash)}</span>
                          <span className="tabular-data">Final cash {formatCurrency(backtestResult.finalCash)}</span>
                          <span className="tabular-data">Final position value {formatCurrency(backtestResult.finalPositionValue)}</span>
                          <span className="tabular-data">Final equity {formatCurrency(backtestResult.finalEquity)}</span>
                          <span>Interval {backtestResult.interval}</span>
                          <span>{backtestResult.startTime.slice(0, 10)} to {backtestResult.endTime.slice(0, 10)}</span>
                        </div>
                      </div>
                      <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                        <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Backtest trades</p>
                        <div className="mt-4 grid gap-3">
                          {backtestResult.orders.length ? backtestResult.orders.map((order) => (
                            <div key={order.id} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.6)] px-4 py-4">
                              <div className="flex items-center justify-between gap-3"><span className="font-semibold">{order.side === "buy" ? "Strategy Buy" : "Strategy Sell"}</span><span className="tabular-data text-sm text-[var(--muted)]">{order.createdAt.slice(0, 10)}</span></div>
                              <p className="tabular-data mt-2 text-sm text-[var(--muted)]">{formatNumber(order.baseQuantity)} {backtestResult.baseAsset} at {formatCurrency(order.expectedPrice)}</p>
                              {order.side === "sell" ? <p className={`tabular-data mt-2 text-sm ${pnlTone(order.realizedPnL)}`}>Realized {formatCurrency(order.realizedPnL)}</p> : null}
                            </div>
                          )) : <EmptyPanelState>No trades matched the current backtest settings.</EmptyPanelState>}
                        </div>
                      </div>
                    </div>
                    <div className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                      <div className="flex items-center justify-between gap-3">
                        <div>
                          <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Recent backtests</p>
                          <p className="mt-2 text-sm text-[var(--muted)]">Use this table to compare the last few read-only runs for the same strategy surface.</p>
                        </div>
                        <span className="text-sm text-[var(--muted)]">{backtestHistory.length} saved</span>
                      </div>
                      <div className="mt-4 overflow-x-auto">
                        <table className="min-w-full text-left text-sm text-[var(--muted)]">
                          <thead>
                            <tr className="border-b border-[var(--line)] text-xs uppercase tracking-[0.22em] text-[var(--muted)]">
                              <th className="pb-3 pr-4 font-medium">Range</th>
                              <th className="pb-3 pr-4 font-medium">Return</th>
                              <th className="pb-3 pr-4 font-medium">Trades</th>
                              <th className="pb-3 pr-4 font-medium">Hit rate</th>
                              <th className="pb-3 font-medium">Drawdown</th>
                            </tr>
                          </thead>
                          <tbody>
                            {backtestHistory.map((entry) => (
                              <tr key={entry.id} className="border-b border-[rgba(255,255,255,0.05)] last:border-b-0">
                                <td className="tabular-data py-3 pr-4">{entry.label}</td>
                                <td className={`tabular-data py-3 pr-4 ${pnlTone(entry.run.summary.returnPercent)}`}>{formatNumber(entry.run.summary.returnPercent)}%</td>
                                <td className="tabular-data py-3 pr-4">{entry.run.summary.tradeCount}</td>
                                <td className="tabular-data py-3 pr-4">{formatNumber(entry.run.summary.hitRatePercent)}%</td>
                                <td className="tabular-data py-3">{formatNumber(entry.run.summary.maxDrawdownPercent)}%</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  </div>
                ) : null}
              </section>
            ) : null}
            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">{detailOnly ? "Trade tickets" : "Balances"}</p>
              {detailOnly ? (
                <div className="mt-5 grid gap-4 md:grid-cols-2">
                  <form className="grid gap-3 rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4" onSubmit={(event) => { event.preventDefault(); void submitOrder("buy"); }}>
                    <p className="text-lg font-semibold">Buy {selectedMarket}</p>
                    <label className="grid gap-2 text-sm text-[var(--muted)]" htmlFor="detail-buy-quote-amount">
                      Buy amount (USDT)
                      <input id="detail-buy-quote-amount" name="detail-buy-quote-amount" aria-label="Buy quote amount" type="number" inputMode="decimal" step="any" autoComplete="off" value={buyQuoteAmount} onChange={(event) => setBuyQuoteAmount(event.target.value)} className={`${baseControlClassName} tabular-data`} />
                    </label>
                    <button type="submit" disabled={isSubmitting || !activeWalletID} className={accentButtonClassName}>{isSubmitting ? "Executing…" : "Run Demo Buy"}</button>
                  </form>
                  <form className="grid gap-3 rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4" onSubmit={(event) => { event.preventDefault(); void submitOrder("sell"); }}>
                    <p className="text-lg font-semibold">Sell {selectedMarket}</p>
                    <label className="grid gap-2 text-sm text-[var(--muted)]" htmlFor="detail-sell-quantity">
                      Sell quantity ({selectedPosition?.baseAsset ?? "base"})
                    <input id="detail-sell-quantity" name="detail-sell-quantity" aria-label="Sell quantity" type="number" inputMode="decimal" step="any" autoComplete="off" value={sellBaseQuantity} onChange={(event) => {
                      sellBaseQuantityRef.current = event.target.value;
                      setSellBaseQuantity(event.target.value);
                    }} className={`${baseControlClassName} tabular-data`} />
                    </label>
                    <button type="button" onClick={() => {
                      if (!selectedPosition) {
                        return;
                      }

                      const nextQuantity = String(selectedPosition.openQuantity);
                      sellBaseQuantityRef.current = nextQuantity;
                      setSellBaseQuantity(nextQuantity);
                    }} disabled={!canUseMaxSell} className={secondaryButtonClassName}>Max position</button>
                    <button type="submit" disabled={isSubmitting || !activeWalletID || !selectedPosition} className={warmButtonClassName}>{isSubmitting ? "Executing…" : "Run Demo Sell"}</button>
                  </form>
                </div>
              ) : (
                <div className="mt-5 grid gap-3">{portfolio?.balances.length ? portfolio.balances.map((balance) => <div key={balance.assetSymbol} className="flex items-center justify-between rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4"><span className="font-medium">{balance.assetSymbol}</span><span className="tabular-data font-[var(--font-mono)] text-[var(--muted)]">{formatNumber(balance.available)}</span></div>) : <EmptyPanelState>Balances will appear here as soon as the session portfolio is available.</EmptyPanelState>}</div>
              )}
            </section>

            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">{detailOnly ? "Current market position" : "Open positions"}</p>
              <div className="mt-5 grid gap-3">
                {(detailOnly ? (selectedPosition ? [selectedPosition] : []) : portfolio?.positions ?? []).length ? (
                  (detailOnly ? (selectedPosition ? [selectedPosition] : []) : portfolio?.positions ?? []).map((position) => (
                    <div key={position.marketSymbol} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                      <div className="flex items-center justify-between gap-4"><span className="text-lg font-semibold">{position.marketSymbol}</span><span className="text-sm text-[var(--accent)]">{formatCurrency(position.positionValue)}</span></div>
                      <p className="tabular-data mt-2 text-sm text-[var(--muted)]">Open qty {formatNumber(position.openQuantity)} at avg {formatCurrency(position.entryPriceAvg)}</p>
                      <div className="mt-3 grid gap-2 text-sm text-[var(--muted)] md:grid-cols-2">
                        <span className="tabular-data">Cost basis {formatCurrency(position.costBasisValue)}</span>
                        <span className="tabular-data">Current {formatCurrency(position.currentPrice)}</span>
                        <span className={`tabular-data ${pnlTone(position.realizedPnL)}`}>Realized {formatCurrency(position.realizedPnL)}</span>
                        <span className={`tabular-data ${pnlTone(position.unrealizedPnL)}`}>Unrealized {formatCurrency(position.unrealizedPnL)}</span>
                      </div>
                    </div>
                  ))
                ) : <EmptyPanelState>No open positions yet.</EmptyPanelState>}
              </div>
            </section>

            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">{detailOnly ? "Selected market order history" : "Recent orders"}</p>
              <div className="mt-5 grid gap-3">
                {visibleOrders.length ? visibleOrders.map((order) => (
                  <div key={order.id} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                    <div className="flex items-center justify-between gap-4"><span className="text-lg font-semibold">{order.marketSymbol}</span><span className={order.side === "sell" ? "text-[var(--accent-warm)]" : "text-[var(--accent)]"}>{order.side}</span></div>
                    <p className="tabular-data mt-2 text-sm text-[var(--muted)]">{formatNumber(order.baseQuantity)} units at {formatCurrency(order.expectedPrice)}</p>
                    <div className="mt-3 flex flex-wrap gap-3 text-sm text-[var(--muted)]"><span className="tabular-data">Quote {formatCurrency(order.quoteAmount)}</span><span className="tabular-data">After trade {formatNumber(order.positionAfter)}</span><span>{order.orderSource === "strategy" ? "Automated" : "Manual"}</span>{order.side === "sell" ? <span className={`tabular-data ${pnlTone(order.realizedPnL)}`}>Realized {formatCurrency(order.realizedPnL)}</span> : null}</div>
                  </div>
                )) : <EmptyPanelState>No orders yet for this scope.</EmptyPanelState>}
              </div>
            </section>

            <section className="rounded-[32px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">{detailOnly ? "Selected market activity" : "Activity log"}</p>
              <div className="mt-5 grid gap-3">
                {activityError ? <div role="status" aria-live="polite" className="rounded-2xl border border-[rgba(255,107,120,0.35)] bg-[rgba(255,107,120,0.08)] px-4 py-4 text-sm text-[var(--accent-hot)]">{activityError}</div> : null}
                {visibleActivity.length ? visibleActivity.map((item) => (
                  <div key={item.id} className="rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.45)] px-4 py-4">
                    <div className="flex items-center justify-between gap-4"><span className="text-lg font-semibold">{item.title}</span><span className="text-sm uppercase text-[var(--accent-hot)]">{item.logType}</span></div>
                    <p className="mt-2 text-sm text-[var(--muted)]">{item.message}</p>
                  </div>
                )) : activityError ? null : <EmptyPanelState>Activity will populate as soon as trades are executed for this scope.</EmptyPanelState>}
              </div>
            </section>
          </div>
        </div>
      </section>
    </main>
  );
}
