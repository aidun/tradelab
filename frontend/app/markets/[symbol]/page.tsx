import React from "react";

import { MarketDashboard } from "@/components/market-dashboard";

type MarketDetailPageProps = {
  params: Promise<{
    symbol: string;
  }>;
};

export default async function MarketDetailPage({ params }: MarketDetailPageProps) {
  const resolvedParams = await params;
  return <MarketDashboard detailOnly initialMarket={decodeURIComponent(resolvedParams.symbol)} />;
}
