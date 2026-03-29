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

export type DemoSession = {
  id: string;
  userID: string;
  walletID: string;
  token: string;
  expiresAt: string;
};

export type RegisteredAccount = {
  userID: string;
  walletID: string;
  clerkUserID: string;
  email: string;
  displayName: string;
  mode: "registered";
};

export type Candle = {
  openTime: string;
  closeTime: string;
  openPrice: number;
  highPrice: number;
  lowPrice: number;
  closePrice: number;
  baseVolume: number;
  quoteVolume: number;
  trades: number;
};

export type MarketDataMeta = {
  source: "fresh" | "stale";
  generatedAt: string;
};

export type CandleFeed = {
  candles: Candle[];
  meta: MarketDataMeta;
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "";

function apiUrl(path: string) {
  if (API_BASE_URL === "") {
    return path;
  }

  return `${API_BASE_URL}${path}`;
}

function authHeaders(token?: string) {
  if (!token) {
    return undefined;
  }

  return {
    Authorization: `Bearer ${token}`
  };
}

export async function createDemoSession(): Promise<DemoSession> {
  const response = await fetch(apiUrl("/api/v1/sessions/demo"), {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    }
  });

  if (!response.ok) {
    throw new Error("Failed to create demo session");
  }

  const data = await response.json();
  return {
    id: data.session.id,
    userID: data.session.user_id,
    walletID: data.session.wallet_id,
    token: data.session.token,
    expiresAt: data.session.expires_at
  };
}

export async function fetchMarkets(): Promise<Market[]> {
  const response = await fetch(apiUrl("/api/v1/markets"), { cache: "no-store" });
  if (!response.ok) {
    throw new Error("Failed to load markets");
  }

  const data = await response.json();
  return data.markets;
}

export async function fetchCandles(marketSymbol: string, interval = "1h", limit = 48): Promise<CandleFeed> {
  const encodedSymbol = encodeURIComponent(marketSymbol);
  const response = await fetch(apiUrl(`/api/v1/markets/${encodedSymbol}/candles?interval=${interval}&limit=${limit}`), {
    cache: "no-store"
  });
  if (!response.ok) {
    throw new Error("Failed to load candles");
  }

  const data = await response.json();
  return {
    candles: data.candles,
    meta: {
      source: data.meta.source,
      generatedAt: data.meta.generated_at ?? data.meta.generatedAt
    }
  };
}

export async function fetchPortfolio(walletID: string, token: string): Promise<PortfolioSummary> {
  const response = await fetch(apiUrl(`/api/v1/portfolios/${walletID}`), {
    cache: "no-store",
    headers: authHeaders(token)
  });

  if (!response.ok) {
    throw new Error("Failed to load portfolio");
  }

  const data = await response.json();
  return data.portfolio;
}

export async function fetchOrders(token: string): Promise<Order[]> {
  const response = await fetch(apiUrl("/api/v1/orders"), {
    cache: "no-store",
    headers: authHeaders(token)
  });
  if (!response.ok) {
    throw new Error("Failed to load orders");
  }

  const data = await response.json();
  return data.orders;
}

export async function fetchActivity(token: string): Promise<ActivityLog[]> {
  const response = await fetch(apiUrl("/api/v1/activity"), {
    cache: "no-store",
    headers: authHeaders(token)
  });
  if (!response.ok) {
    throw new Error("Failed to load activity");
  }

  const data = await response.json();
  return data.activity;
}

export async function placeMarketBuy(input: {
  marketSymbol: string;
  quoteAmount: number;
  token: string;
}) {
  const response = await fetch(apiUrl("/api/v1/orders"), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(input.token)
    },
    body: JSON.stringify({
      market_symbol: input.marketSymbol,
      quote_amount: input.quoteAmount
    })
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => ({ error: "Order failed" }));
    throw new Error(payload.error ?? "Order failed");
  }

  return response.json();
}

export async function bootstrapRegisteredAccount(token: string): Promise<RegisteredAccount> {
  const response = await fetch(apiUrl("/api/v1/account/bootstrap"), {
    method: "POST",
    headers: authHeaders(token)
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => ({ error: "Failed to bootstrap account" }));
    throw new Error(payload.error ?? "Failed to bootstrap account");
  }

  const data = await response.json();
  return {
    userID: data.account.user_id,
    walletID: data.account.wallet_id,
    clerkUserID: data.account.clerk_user_id,
    email: data.account.email ?? "",
    displayName: data.account.display_name,
    mode: "registered"
  };
}

export async function upgradeGuestAccount(input: {
  registeredToken: string;
  guestToken: string;
  preserveGuestData: boolean;
}): Promise<RegisteredAccount> {
  const response = await fetch(apiUrl("/api/v1/account/upgrade"), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-TradeLab-Guest-Token": input.guestToken,
      ...authHeaders(input.registeredToken)
    },
    body: JSON.stringify({
      preserve_guest_data: input.preserveGuestData
    })
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => ({ error: "Failed to upgrade guest account" }));
    throw new Error(payload.error ?? "Failed to upgrade guest account");
  }

  const data = await response.json();
  return {
    userID: data.account.user_id,
    walletID: data.account.wallet_id,
    clerkUserID: data.account.clerk_user_id,
    email: data.account.email ?? "",
    displayName: data.account.display_name,
    mode: "registered"
  };
}
