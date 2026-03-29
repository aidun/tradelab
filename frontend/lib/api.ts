export type Market = {
  id: string;
  symbol: string;
  baseAsset: string;
  quoteAsset: string;
  minNotional: number;
  exchange: string;
};

export type Balance = {
  walletID: string;
  assetSymbol: string;
  available: number;
};

export type Position = {
  id: string;
  userID: string;
  walletID: string;
  marketID: string;
  marketSymbol: string;
  baseAsset: string;
  quoteAsset: string;
  status: string;
  entryQuantity: number;
  entryPriceAvg: number;
  currentPrice: number;
  positionValue: number;
  unrealizedPnL: number;
  openedAt: string;
};

export type PortfolioSummary = {
  walletID: string;
  baseCurrency: string;
  totalValue: number;
  cashBalance: number;
  positions: Position[];
  balances: Balance[];
};

export type Order = {
  id: string;
  walletID: string;
  marketSymbol: string;
  quoteAmount: number;
  expectedPrice: number;
  status: string;
  createdAt: string;
};

export type ActivityLog = {
  id: string;
  walletID: string;
  logType: string;
  title: string;
  message: string;
  createdAt: string;
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export async function fetchMarkets(): Promise<Market[]> {
  const response = await fetch(`${API_BASE_URL}/api/v1/markets`, { cache: "no-store" });
  if (!response.ok) {
    throw new Error("Failed to load markets");
  }

  const data = await response.json();
  return data.markets;
}

export async function fetchPortfolio(walletID: string): Promise<PortfolioSummary> {
  const response = await fetch(`${API_BASE_URL}/api/v1/portfolios/${walletID}`, {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error("Failed to load portfolio");
  }

  const data = await response.json();
  return data.portfolio;
}

export async function fetchOrders(walletID: string): Promise<Order[]> {
  const response = await fetch(`${API_BASE_URL}/api/v1/orders?wallet_id=${walletID}`, {
    cache: "no-store"
  });
  if (!response.ok) {
    throw new Error("Failed to load orders");
  }

  const data = await response.json();
  return data.orders;
}

export async function fetchActivity(walletID: string): Promise<ActivityLog[]> {
  const response = await fetch(`${API_BASE_URL}/api/v1/activity?wallet_id=${walletID}`, {
    cache: "no-store"
  });
  if (!response.ok) {
    throw new Error("Failed to load activity");
  }

  const data = await response.json();
  return data.activity;
}

export async function placeMarketBuy(input: {
  userID: string;
  walletID: string;
  marketSymbol: string;
  quoteAmount: number;
  expectedPrice: number;
}) {
  const response = await fetch(`${API_BASE_URL}/api/v1/orders`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      user_id: input.userID,
      wallet_id: input.walletID,
      market_symbol: input.marketSymbol,
      quote_amount: input.quoteAmount,
      expected_price: input.expectedPrice
    })
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => ({ error: "Order failed" }));
    throw new Error(payload.error ?? "Order failed");
  }

  return response.json();
}
