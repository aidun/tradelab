"use client";

import React, { useMemo, useState } from "react";

import { MarketDashboard } from "@/components/market-dashboard";
import { AuthGateActions, useTradeLabAuth } from "@/lib/tradelab-auth";

type TradeLabAppShellProps = {
  detailOnly?: boolean;
  initialMarket?: string;
  requestedPath?: string;
};

function LoadingWorkspaceShell({ message }: { message: string }) {
  return (
    <main className="grid-glow min-h-screen overflow-hidden px-6 py-8 md:px-10 lg:px-14">
      <section className="mx-auto flex min-h-[80vh] w-full max-w-7xl items-center justify-center">
        <div className="w-full max-w-2xl rounded-[32px] border border-[var(--line)] bg-[rgba(7,17,31,0.88)] p-10 text-center shadow-[0_30px_90px_rgba(0,0,0,0.35)]">
          <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.32em] text-[var(--accent)]">TradeLab</p>
          <h1 className="mt-5 text-4xl font-semibold tracking-[-0.05em] text-balance text-[var(--text)]">Preparing your workspace</h1>
          <p className="mt-4 text-base leading-7 text-[var(--muted)]">{message}</p>
        </div>
      </section>
    </main>
  );
}

function EntryGate({
  canUseAuth,
  onContinueAsGuest,
  requestedPath
}: {
  canUseAuth: boolean;
  onContinueAsGuest: () => void;
  requestedPath: string;
}) {
  const destinationLabel =
    requestedPath !== "/"
      ? "Your requested market route will open immediately after access is established."
      : "Choose how you want to access the trading workspace.";

  return (
    <main className="grid-glow min-h-screen overflow-hidden px-6 py-8 md:px-10 lg:px-14">
      <section className="mx-auto grid min-h-[80vh] w-full max-w-7xl gap-6 lg:grid-cols-[1.2fr_0.8fr]">
        <section className="rounded-[36px] border border-[var(--line)] bg-[linear-gradient(180deg,rgba(8,17,14,0.98),rgba(5,11,9,0.98))] p-8 shadow-[0_30px_90px_rgba(0,0,0,0.35)] md:p-10">
          <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.32em] text-[var(--accent)]">TradeLab</p>
          <h1 className="mt-6 max-w-3xl text-5xl font-semibold leading-none tracking-[-0.05em] text-balance text-[var(--text)] md:text-7xl">Access your trading workspace.</h1>
          <p className="mt-6 max-w-2xl text-lg leading-8 text-[var(--muted)]">A professional crypto analytics and execution surface with deliberate account access, compact market hierarchy, and optional guest entry.</p>

          <div className="mt-12 grid gap-4 md:grid-cols-3">
            <div className="rounded-[28px] border border-[var(--line)] bg-[rgba(9,20,16,0.82)] p-5">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Market</p>
              <p className="mt-4 text-2xl font-semibold text-[var(--text)]">XRP / USDT</p>
              <p className="tabular-data mt-3 text-2xl font-semibold text-[var(--accent)]">0.7345</p>
              <p className="mt-4 text-sm leading-6 text-[var(--muted)]">Tight market context, stronger number hierarchy, and less dashboard noise.</p>
            </div>
            <div className="rounded-[28px] border border-[var(--line)] bg-[rgba(9,20,16,0.82)] p-5">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Execution</p>
              <p className="mt-4 text-2xl font-semibold text-[var(--text)]">Place order</p>
              <p className="mt-3 text-sm leading-6 text-[var(--muted)]">Execution stays compact, local, and confidence-oriented instead of tutorial-driven.</p>
            </div>
            <div className="rounded-[28px] border border-[var(--line)] bg-[rgba(9,20,16,0.82)] p-5">
              <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.24em] text-[var(--muted)]">Trust</p>
              <p className="mt-4 text-2xl font-semibold text-[var(--text)]">One-time disclosure</p>
              <p className="mt-3 text-sm leading-6 text-[var(--muted)]">Financial boundaries are acknowledged once, then moved out of the primary product flow.</p>
            </div>
          </div>
        </section>

        <aside className="rounded-[36px] border border-[var(--line)] bg-[rgba(7,17,31,0.9)] p-8 shadow-[0_30px_90px_rgba(0,0,0,0.35)] md:p-10">
          <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.32em] text-[var(--accent)]">Entry</p>
          <h2 className="mt-6 text-3xl font-semibold tracking-[-0.04em] text-[var(--text)]">Choose your path</h2>
          <p className="mt-4 text-base leading-7 text-[var(--muted)]">{destinationLabel}</p>

          {canUseAuth ? (
            <div className="mt-8">
              <AuthGateActions />
            </div>
          ) : null}

          <div className="mt-4">
            <button
              type="button"
              onClick={onContinueAsGuest}
              className="focus-ring w-full rounded-2xl border border-[var(--line)] bg-[rgba(7,17,31,0.58)] px-4 py-3 text-sm font-medium text-[var(--text)] transition hover:border-[var(--accent)] hover:text-[var(--accent)]"
            >
              Continue as guest
            </button>
          </div>

          <div className="mt-8 rounded-[24px] border border-[var(--line)] bg-[rgba(9,20,16,0.64)] p-5">
            <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.22em] text-[var(--muted)]">Access note</p>
            <p className="mt-3 text-sm leading-7 text-[var(--muted)]">Guest entry is intentional and immediate. Account creation remains the primary path for durable access and later continuity.</p>
          </div>
        </aside>
      </section>
    </main>
  );
}

/** TradeLabAppShell decides between the v1 entry gate and the trading workspace. */
export function TradeLabAppShell({
  detailOnly = false,
  initialMarket = "XRP/USDT",
  requestedPath = "/"
}: TradeLabAppShellProps) {
  const auth = useTradeLabAuth();
  const [guestAccessRequested, setGuestAccessRequested] = useState(false);

  const canUseAuth = useMemo(
    () => auth.available && (auth.provider === "mock" || auth.provider === "clerk"),
    [auth.available, auth.provider]
  );

  if (auth.status === "loading") {
    return <LoadingWorkspaceShell message="Opening the workspace and loading the current portfolio." />;
  }

  if (auth.status === "signed_out" && !guestAccessRequested) {
    return (
      <EntryGate
        canUseAuth={canUseAuth}
        onContinueAsGuest={() => setGuestAccessRequested(true)}
        requestedPath={requestedPath}
      />
    );
  }

  return (
    <MarketDashboard
      detailOnly={detailOnly}
      initialMarket={initialMarket}
      autoStartGuest={guestAccessRequested}
    />
  );
}
