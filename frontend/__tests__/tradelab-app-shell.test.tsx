import React from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { TradeLabAppShell } from "@/components/tradelab-app-shell";

const {
  mockUseTradeLabAuth,
  mockCreateDemoSession,
  mockMarketDashboard
} = vi.hoisted(() => ({
  mockUseTradeLabAuth: vi.fn(),
  mockCreateDemoSession: vi.fn(),
  mockMarketDashboard: vi.fn()
}));

vi.mock("@/components/market-dashboard", () => ({
  MarketDashboard: (props: { detailOnly?: boolean; initialMarket?: string; autoStartGuest?: boolean }) => {
    mockMarketDashboard(props);
    return (
      <div data-testid="workspace">
        {props.detailOnly ? `detail:${props.initialMarket}` : `workspace:${props.autoStartGuest ? "guest" : "account"}`}
      </div>
    );
  }
}));

vi.mock("@/lib/tradelab-auth", () => ({
  useTradeLabAuth: mockUseTradeLabAuth,
  AuthGateActions: () => <div>Auth actions</div>
}));

vi.mock("@/lib/api", () => ({
  createDemoSession: mockCreateDemoSession
}));

describe("TradeLabAppShell disclaimer flow", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
    window.sessionStorage.clear();
    mockUseTradeLabAuth.mockReturnValue({
      available: true,
      provider: "mock",
      status: "signed_out",
      user: null
    });
    mockCreateDemoSession.mockResolvedValue({
      id: "session-1",
      userID: "user-1",
      walletID: "wallet-1",
      token: "guest-token",
      expiresAt: "2026-04-29T12:00:00Z"
    });
  });

  it("shows the disclaimer once for first-time guest access before the workspace opens", async () => {
    render(<TradeLabAppShell requestedPath="/" />);

    fireEvent.click(screen.getByRole("button", { name: /continue as guest/i }));

    expect(await screen.findByRole("heading", { name: /financial boundary/i })).toBeInTheDocument();
    expect(screen.getByText(/simulated trading only/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /acknowledge and continue/i }));

    await waitFor(() => {
      expect(screen.getByTestId("workspace")).toHaveTextContent("workspace:guest");
    });
  });

  it("skips the disclaimer for the same guest identity after acknowledgement", async () => {
    window.sessionStorage.setItem(
      "tradelab.demo-session",
      JSON.stringify({
        id: "session-1",
        userID: "user-1",
        walletID: "wallet-1",
        token: "guest-token",
        expiresAt: "2026-04-29T12:00:00Z"
      })
    );
    window.sessionStorage.setItem("tradelab.disclaimer.guest.wallet-1", "acknowledged");

    render(<TradeLabAppShell requestedPath="/markets/XRP%2FUSDT" detailOnly initialMarket="XRP/USDT" />);

    fireEvent.click(screen.getByRole("button", { name: /continue as guest/i }));

    await waitFor(() => {
      expect(screen.getByTestId("workspace")).toHaveTextContent("detail:XRP/USDT");
    });

    expect(screen.queryByRole("heading", { name: /financial boundary/i })).not.toBeInTheDocument();
    expect(mockCreateDemoSession).not.toHaveBeenCalled();
  });

  it("shows the disclaimer once for signed-in account access and skips it after acknowledgement", async () => {
    mockUseTradeLabAuth.mockReturnValue({
      available: true,
      provider: "mock",
      status: "signed_in",
      user: {
        clerkUserID: "account-1",
        email: "trader@example.com",
        displayName: "Trader Example"
      }
    });

    const { rerender } = render(<TradeLabAppShell requestedPath="/" />);

    expect(screen.getByRole("heading", { name: /financial boundary/i })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /acknowledge and continue/i }));

    await waitFor(() => {
      expect(screen.getByTestId("workspace")).toHaveTextContent("workspace:account");
    });

    rerender(<TradeLabAppShell requestedPath="/" />);

    await waitFor(() => {
      expect(screen.getByTestId("workspace")).toHaveTextContent("workspace:account");
    });

    expect(screen.queryByRole("heading", { name: /financial boundary/i })).not.toBeInTheDocument();
  });
});
