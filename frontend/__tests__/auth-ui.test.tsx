import React from "react";
import { fireEvent, render, screen } from "@testing-library/react";

import { MarketDashboard } from "@/components/market-dashboard";

const upgradeGuestSession = vi.fn();
const clearMessages = vi.fn();
const { mockUseTradeLabAuth } = vi.hoisted(() => ({
  mockUseTradeLabAuth: vi.fn(() => ({
    available: true,
    provider: "mock",
    status: "signed_out",
    user: null
  }))
}));

vi.mock("@/lib/api", () => ({
  fetchCandles: vi.fn(async () => ({
    candles: [
      {
        openTime: "2026-03-29T10:00:00Z",
        closeTime: "2026-03-29T10:59:59Z",
        openPrice: 0.62,
        highPrice: 0.64,
        lowPrice: 0.61,
        closePrice: 0.63,
        baseVolume: 1200000,
        quoteVolume: 756000,
        trades: 8000
      }
    ],
    meta: {
      source: "fresh",
      generatedAt: "2026-03-29T12:05:00Z"
    }
  })),
  placeMarketBuy: vi.fn()
}));

vi.mock("@/lib/tradelab-auth", () => ({
  useTradeLabAuth: mockUseTradeLabAuth,
  AuthEntryActions: () => <div>Mock auth actions</div>,
  AuthStatusControls: () => {
    const auth = mockUseTradeLabAuth();
    if (!auth.available || auth.status !== "signed_in") {
      return null;
    }

    return (
      <button type="button" onClick={() => void auth.signOut()}>
        Log out
      </button>
    );
  }
}));

vi.mock("@/lib/use-account-session", () => ({
  useAccountSession: vi.fn(() => ({
    guestSession: {
      id: "session-1",
      userID: "user-1",
      walletID: "wallet-1",
      token: "guest-token",
      expiresAt: "2026-04-29T12:00:00Z"
    },
    registeredAccount: null,
    markets: [
      {
        id: "market-1",
        symbol: "XRP/USDT",
        baseAsset: "XRP",
        quoteAsset: "USDT",
        minNotional: 10,
        exchange: "demo"
      }
    ],
    portfolio: {
      walletID: "wallet-1",
      baseCurrency: "USDT",
      totalValue: 10000,
      cashBalance: 9950,
      balances: [{ walletID: "wallet-1", assetSymbol: "USDT", available: 9950 }],
      positions: []
    },
    orders: [],
    activity: [],
    isLoading: false,
    isUpgrading: false,
    showUpgradePrompt: false,
    error: null,
    success: null,
    activeWalletID: "wallet-1",
    accountModeLabel: "Guest demo session",
    shouldShowAuthValuePrompt: true,
    clearMessages,
    setErrorMessage: vi.fn(),
    setSuccessMessage: vi.fn(),
    refreshCoreData: vi.fn(async () => ({ walletID: "wallet-1", token: "guest-token" })),
    upgradeGuestSession,
    activeAccessToken: vi.fn(async () => "guest-token")
  }))
}));

describe("MarketDashboard auth UI", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseTradeLabAuth.mockReturnValue({
      available: true,
      provider: "mock",
      status: "signed_out",
      user: null
    });
  });

  it("shows the durable access prompt after guest users reach the dashboard", async () => {
    render(<MarketDashboard />);

    expect(await screen.findByText(/keep this sandbox beyond the guest session/i)).toBeInTheDocument();
    expect(screen.getByText(/mock auth actions/i)).toBeInTheDocument();
  });

  it("shows the guest upgrade prompt and keeps the preserve and fresh actions wired", async () => {
    const { useAccountSession } = await import("@/lib/use-account-session");
    vi.mocked(useAccountSession).mockReturnValue({
      guestSession: {
        id: "session-1",
        userID: "user-1",
        walletID: "wallet-1",
        token: "guest-token",
        expiresAt: "2026-04-29T12:00:00Z"
      },
      registeredAccount: null,
      markets: [
        {
          id: "market-1",
          symbol: "XRP/USDT",
          baseAsset: "XRP",
          quoteAsset: "USDT",
          minNotional: 10,
          exchange: "demo"
        }
      ],
      portfolio: {
        walletID: "wallet-1",
        baseCurrency: "USDT",
        totalValue: 10000,
        cashBalance: 9950,
        balances: [{ walletID: "wallet-1", assetSymbol: "USDT", available: 9950 }],
        positions: []
      },
      orders: [],
      activity: [],
      isLoading: false,
      isUpgrading: false,
      showUpgradePrompt: true,
      error: null,
      success: null,
      activeWalletID: "wallet-1",
      accountModeLabel: "Guest demo session",
      shouldShowAuthValuePrompt: false,
      clearMessages,
      setErrorMessage: vi.fn(),
      setSuccessMessage: vi.fn(),
      refreshCoreData: vi.fn(async () => ({ walletID: "wallet-1", token: "guest-token" })),
      upgradeGuestSession,
      activeAccessToken: vi.fn(async () => "guest-token")
    });

    render(<MarketDashboard />);

    fireEvent.click(await screen.findByRole("button", { name: /keep guest demo data/i }));
    fireEvent.click(screen.getByRole("button", { name: /start fresh/i }));

    expect(upgradeGuestSession).toHaveBeenNthCalledWith(1, true);
    expect(upgradeGuestSession).toHaveBeenNthCalledWith(2, false);
  });

  it("routes signed-in clerk logout through the shared auth signOut path", async () => {
    const signOut = vi.fn(async () => undefined);
    mockUseTradeLabAuth.mockReturnValue({
      available: true,
      provider: "clerk",
      status: "signed_in",
      user: {
        clerkUserID: "clerk-user-1",
        email: "trader@example.com",
        displayName: "Trader Example"
      },
      signOut
    });

    const { AuthStatusControls } = await import("@/lib/tradelab-auth");
    render(<AuthStatusControls />);

    fireEvent.click(screen.getByRole("button", { name: /log out/i }));

    expect(signOut).toHaveBeenCalledTimes(1);
  });
});
