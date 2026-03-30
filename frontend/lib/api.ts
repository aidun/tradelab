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
  openQuantity: number;
  entryQuantity: number;
  entryPriceAvg: number;
  currentPrice: number;
  costBasisValue: number;
  positionValue: number;
  realizedPnL: number;
  unrealizedPnL: number;
  openedAt: string;
};

export type PortfolioAllocation = {
  marketSymbol: string;
  value: number;
  weight: number;
};

export type AccountingMode = "average_cost" | "fifo" | "hybrid";

export type PortfolioSummary = {
  walletID: string;
  baseCurrency: string;
  accountingMode: AccountingMode;
  totalValue: number;
  cashBalance: number;
  positionValue: number;
  realizedPnL: number;
  unrealizedPnL: number;
  positions: Position[];
  balances: Balance[];
  allocations: PortfolioAllocation[];
};

export type Order = {
  id: string;
  walletID: string;
  marketSymbol: string;
  strategyID?: string;
  orderSource: "manual" | "strategy" | "system";
  side: "buy" | "sell";
  baseQuantity: number;
  quoteAmount: number;
  expectedPrice: number;
  status: string;
  realizedPnL: number;
  positionAfter: number;
  createdAt: string;
};

export type ActivityLog = {
  id: string;
  walletID: string;
  marketSymbol: string;
  logType: string;
  title: string;
  message: string;
  createdAt: string;
};

export type StrategyStatus = "draft" | "active" | "paused" | "archived";

export type StrategyConfig = {
  dipBuy: {
    enabled: boolean;
    dipPercent: number;
    spendQuoteAmount: number;
  };
  takeProfit: {
    enabled: boolean;
    triggerPercent: number;
  };
  stopLoss: {
    enabled: boolean;
    triggerPercent: number;
  };
};

export type Strategy = {
  id: string;
  walletID: string;
  marketSymbol: string;
  status: StrategyStatus;
  config: StrategyConfig;
  referencePrice: number;
  lastRunAt?: string;
  lastDecision: "buy" | "sell" | "hold" | "";
  lastOutcome: "executed" | "skipped" | "errored" | "";
  lastReason: string;
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

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "";
const UI_ERROR_RUNTIME = process.env.NODE_ENV ?? "development";

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

async function parseApiError(response: Response, fallback: string) {
  const payload = await response.json().catch(() => ({ error: fallback }));
  throw new ApiError(resolveApiErrorMessage(payload.error, fallback), response.status);
}

export function resolveApiErrorMessage(detail: unknown, fallback: string, runtime = UI_ERROR_RUNTIME) {
  const detailMessage = typeof detail === "string" ? detail.trim() : "";
  if (runtime !== "production") {
    return detailMessage || fallback;
  }

  if (detailMessage !== "") {
    console.error("TradeLab API error", detailMessage);
  }

  return fallback;
}

export async function createDemoSession(): Promise<DemoSession> {
  const response = await fetch(apiUrl("/api/v1/sessions/demo"), {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    }
  });

  if (!response.ok) {
    await parseApiError(response, "We couldn't start a demo session right now.");
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
    await parseApiError(response, "We couldn't load market data right now.");
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
    await parseApiError(response, "We couldn't refresh the chart right now.");
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

export async function fetchPortfolio(walletID: string, token: string, accountingMode: AccountingMode): Promise<PortfolioSummary> {
  const response = await fetch(apiUrl(`/api/v1/portfolios/${walletID}?accounting_mode=${accountingMode}`), {
    cache: "no-store",
    credentials: "include",
    headers: authHeaders(token)
  });

  if (!response.ok) {
    await parseApiError(response, "We couldn't load the portfolio right now.");
  }

  const data = await response.json();
  return data.portfolio;
}

export async function fetchOrders(
  token: string,
  options?: { accountingMode?: AccountingMode; marketSymbol?: string }
): Promise<Order[]> {
  const params = new URLSearchParams();
  if (options?.accountingMode) {
    params.set("accounting_mode", options.accountingMode);
  }
  if (options?.marketSymbol) {
    params.set("market_symbol", options.marketSymbol);
  }

  const response = await fetch(apiUrl(`/api/v1/orders${params.size > 0 ? `?${params.toString()}` : ""}`), {
    cache: "no-store",
    credentials: "include",
    headers: authHeaders(token)
  });
  if (!response.ok) {
    await parseApiError(response, "We couldn't load recent orders right now.");
  }

  const data = await response.json();
  return data.orders;
}

export async function fetchActivity(token: string, options?: { marketSymbol?: string }): Promise<ActivityLog[]> {
  const params = new URLSearchParams();
  if (options?.marketSymbol) {
    params.set("market_symbol", options.marketSymbol);
  }

  const response = await fetch(apiUrl(`/api/v1/activity${params.size > 0 ? `?${params.toString()}` : ""}`), {
    cache: "no-store",
    credentials: "include",
    headers: authHeaders(token)
  });
  if (!response.ok) {
    await parseApiError(response, "We couldn't load activity right now.");
  }

  const data = await response.json();
  return data.activity;
}

export async function fetchStrategies(token: string, marketSymbol?: string): Promise<Strategy[]> {
  const params = new URLSearchParams();
  if (marketSymbol) {
    params.set("market_symbol", marketSymbol);
  }

  const response = await fetch(apiUrl(`/api/v1/strategies${params.size > 0 ? `?${params.toString()}` : ""}`), {
    cache: "no-store",
    credentials: "include",
    headers: authHeaders(token)
  });
  if (!response.ok) {
    await parseApiError(response, "We couldn't load automation right now.");
  }

  const data = await response.json();
  return data.strategies;
}

export async function saveStrategy(input: {
  marketSymbol: string;
  status: StrategyStatus;
  config: StrategyConfig;
  token?: string | null;
}): Promise<Strategy> {
  const response = await fetch(apiUrl("/api/v1/strategies"), {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(input.token ?? undefined)
    },
    body: JSON.stringify({
      market_symbol: input.marketSymbol,
      status: input.status,
      config: input.config
    })
  });

  if (!response.ok) {
    await parseApiError(response, "We couldn't save the automation changes.");
  }

  const data = await response.json();
  return data.strategy;
}

export async function patchStrategy(input: {
  id: string;
  status: StrategyStatus;
  config: StrategyConfig;
  token?: string | null;
}): Promise<Strategy> {
  const response = await fetch(apiUrl(`/api/v1/strategies/${input.id}`), {
    method: "PATCH",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(input.token ?? undefined)
    },
    body: JSON.stringify({
      status: input.status,
      config: input.config
    })
  });

  if (!response.ok) {
    await parseApiError(response, "We couldn't update the automation changes.");
  }

  const data = await response.json();
  return data.strategy;
}

export async function placeMarketOrder(input: {
  side: "buy" | "sell";
  marketSymbol: string;
  quoteAmount?: number;
  baseQuantity?: number;
  token?: string | null;
}) {
  const response = await fetch(apiUrl("/api/v1/orders"), {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(input.token ?? undefined)
    },
    body: JSON.stringify({
      market_symbol: input.marketSymbol,
      side: input.side,
      quote_amount: input.quoteAmount,
      base_quantity: input.baseQuantity
    })
  });

  if (!response.ok) {
    await parseApiError(response, "We couldn't place that order right now.");
  }

  return response.json();
}

export async function bootstrapRegisteredAccount(token: string): Promise<RegisteredAccount> {
  const response = await fetch(apiUrl("/api/v1/account/bootstrap"), {
    method: "POST",
    credentials: "include",
    headers: authHeaders(token)
  });

  if (!response.ok) {
    await parseApiError(response, "We couldn't open the registered account right now.");
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
    credentials: "include",
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
    await parseApiError(response, "We couldn't upgrade this guest session right now.");
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

export async function logoutRegisteredAccount(): Promise<void> {
  const response = await fetch(apiUrl("/api/v1/account/logout"), {
    method: "POST",
    credentials: "include"
  });

  if (!response.ok) {
    await parseApiError(response, "We couldn't log out cleanly right now.");
  }
}
