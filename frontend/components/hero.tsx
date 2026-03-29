import React from "react";
import { MetricCard } from "@/components/metric-card";

const metrics = [
  {
    label: "Reference market",
    value: "XRP / USDT",
    trend: "Fast onboarding flow, multi-asset architecture"
  },
  {
    label: "Strategy mode",
    value: "Auto + Manual",
    trend: "Built for paper trading first"
  },
  {
    label: "Quality bar",
    value: "Tests required",
    trend: "Backend, UI, and end-to-end coverage"
  }
];

export function Hero() {
  return (
    <main className="grid-glow min-h-screen overflow-hidden px-6 py-8 md:px-10 lg:px-14">
      <section className="mx-auto flex w-full max-w-7xl flex-col gap-8">
        <div className="rounded-[40px] border border-[var(--line)] bg-[var(--surface-strong)] p-6 shadow-[0_30px_90px_rgba(0,0,0,0.35)] backdrop-blur md:p-8">
          <div className="flex flex-col gap-12 lg:flex-row lg:items-end lg:justify-between">
            <div className="max-w-3xl">
              <p className="font-[var(--font-mono)] text-sm uppercase tracking-[0.32em] text-[var(--accent)]">
                TradeLab
              </p>
              <h1 className="mt-4 text-5xl font-semibold leading-none tracking-[-0.05em] md:text-7xl">
                Paper trading with
                <span className="block text-[var(--accent-warm)]">real strategy discipline.</span>
              </h1>
              <p className="mt-6 max-w-2xl text-lg leading-8 text-[var(--muted)]">
                Build, test, and review automated crypto strategies in a premium demo environment.
                XRP is the starting point, not the limit.
              </p>
            </div>

            <div className="rounded-[28px] border border-[var(--line)] bg-[rgba(7,17,31,0.78)] p-5 font-[var(--font-mono)] text-sm text-[var(--muted)]">
              <div className="flex items-center justify-between gap-8">
                <span>Simulation status</span>
                <span className="text-[var(--accent)]">Active</span>
              </div>
              <div className="mt-4 h-px bg-[var(--line)]" />
              <div className="mt-4 flex items-center justify-between gap-8">
                <span>Execution model</span>
                <span>Market + strategy</span>
              </div>
              <div className="mt-4 flex items-center justify-between gap-8">
                <span>Storage target</span>
                <span>PostgreSQL</span>
              </div>
            </div>
          </div>

          <div className="mt-10 grid gap-4 md:grid-cols-3">
            {metrics.map((metric) => (
              <MetricCard key={metric.label} {...metric} />
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}
