import React from "react";

import { TradeLabAppShell } from "@/components/tradelab-app-shell";

type MarketDetailPageProps = {
  params: Promise<{
    symbol: string;
  }>;
};

export default async function MarketDetailPage({ params }: MarketDetailPageProps) {
  const resolvedParams = await params;
  const decodedSymbol = decodeURIComponent(resolvedParams.symbol);
  return (
    <TradeLabAppShell
      detailOnly
      initialMarket={decodedSymbol}
      requestedPath={`/markets/${encodeURIComponent(decodedSymbol)}`}
    />
  );
}
