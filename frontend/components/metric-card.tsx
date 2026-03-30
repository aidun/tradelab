import React from "react";

type MetricCardProps = {
  label: string;
  value: string;
  trend: string;
};

/** MetricCard renders one compact portfolio metric in the overview header. */
export function MetricCard({ label, value, trend }: MetricCardProps) {
  return (
    <div className="rounded-[28px] border border-[var(--line)] bg-[var(--surface)] p-5 backdrop-blur">
      <p className="font-[var(--font-mono)] text-xs uppercase tracking-[0.28em] text-[var(--muted)]">
        {label}
      </p>
      <p className="mt-3 text-3xl font-semibold">{value}</p>
      <p className="mt-2 text-sm text-[var(--accent)]">{trend}</p>
    </div>
  );
}
