"use client";

import { startTransition, useEffect, useMemo, useRef, useState } from "react";

import {
  ApiError,
  bootstrapRegisteredAccount,
  createDemoSession,
  fetchActivity,
  fetchMarkets,
  fetchOrders,
  fetchPortfolio,
  upgradeGuestAccount,
  type ActivityLog,
  type DemoSession,
  type Market,
  type Order,
  type PortfolioSummary,
  type RegisteredAccount
} from "@/lib/api";
import { useTradeLabAuth } from "@/lib/tradelab-auth";

const DEMO_SESSION_STORAGE_KEY = "tradelab.demo-session";

type CoreDataState = {
  guestSession: DemoSession | null;
  registeredAccount: RegisteredAccount | null;
  markets: Market[];
  portfolio: PortfolioSummary | null;
  orders: Order[];
  activity: ActivityLog[];
  isLoading: boolean;
  isUpgrading: boolean;
  showUpgradePrompt: boolean;
  error: string | null;
  success: string | null;
  activeWalletID: string | null;
  accountModeLabel: string;
  shouldShowAuthValuePrompt: boolean;
  clearMessages: () => void;
  setErrorMessage: (message: string | null) => void;
  setSuccessMessage: (message: string | null) => void;
  refreshCoreData: () => Promise<{ walletID: string; token: string | null } | null>;
  upgradeGuestSession: (preserveGuestData: boolean) => Promise<void>;
  activeAccessToken: () => Promise<string | null>;
};

export function useAccountSession(): CoreDataState {
  const auth = useTradeLabAuth();
  const previousAuthStatus = useRef(auth.status);

  const [guestSession, setGuestSession] = useState<DemoSession | null>(null);
  const [registeredAccount, setRegisteredAccount] = useState<RegisteredAccount | null>(null);
  const [markets, setMarkets] = useState<Market[]>([]);
  const [portfolio, setPortfolio] = useState<PortfolioSummary | null>(null);
  const [orders, setOrders] = useState<Order[]>([]);
  const [activity, setActivity] = useState<ActivityLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isUpgrading, setIsUpgrading] = useState(false);
  const [showUpgradePrompt, setShowUpgradePrompt] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const activeWalletID = registeredAccount?.walletID ?? guestSession?.walletID ?? null;
  const accountModeLabel = registeredAccount ? "Registered demo account" : "Guest demo session";
  const shouldShowAuthValuePrompt = auth.available && auth.status === "signed_out" && !isLoading && portfolio !== null;

  async function ensureGuestSession() {
    if (typeof window === "undefined") {
      throw new Error("Demo session can only be created in the browser");
    }

    const cachedValue = window.sessionStorage.getItem(DEMO_SESSION_STORAGE_KEY);
    if (cachedValue) {
      try {
        const cachedSession = JSON.parse(cachedValue) as DemoSession;
        if (new Date(cachedSession.expiresAt).getTime() > Date.now()) {
          return cachedSession;
        }
      } catch {
        window.sessionStorage.removeItem(DEMO_SESSION_STORAGE_KEY);
      }
    }

    const nextSession = await createDemoSession();
    window.sessionStorage.setItem(DEMO_SESSION_STORAGE_KEY, JSON.stringify(nextSession));
    return nextSession;
  }

  function clearGuestSession() {
    if (typeof window !== "undefined") {
      window.sessionStorage.removeItem(DEMO_SESSION_STORAGE_KEY);
    }
    setGuestSession(null);
  }

  async function loadCoreData(walletID: string, token?: string | null) {
    setError(null);

    const [marketList, portfolioSummary, orderHistory, activityHistory] = await Promise.all([
      fetchMarkets(),
      fetchPortfolio(walletID, token ?? ""),
      fetchOrders(token ?? ""),
      fetchActivity(token ?? "")
    ]);

    setMarkets(marketList);
    setPortfolio(portfolioSummary);
    setOrders(orderHistory);
    setActivity(activityHistory);
  }

  async function bootstrapRegistered() {
    const token = await auth.getToken();
    if (!token) {
      throw new Error("Registered session token is unavailable");
    }

    const account = await bootstrapRegisteredAccount(token);
    setRegisteredAccount(account);
    await loadCoreData(account.walletID);
    return { walletID: account.walletID, token: null };
  }

  async function activateGuestExperience() {
    const session = await ensureGuestSession();
    setGuestSession(session);
    setRegisteredAccount(null);
    await loadCoreData(session.walletID, session.token);
    return { walletID: session.walletID, token: session.token };
  }

  useEffect(() => {
    let cancelled = false;

    if (auth.status === "loading") {
      return;
    }

    setIsLoading(true);

    startTransition(() => {
      const run = async () => {
        if (auth.status === "signed_in" && !guestSession && !registeredAccount) {
          await bootstrapRegistered();
          return;
        }

        // Signed-out users should always fall back to a guest session, but only once the
        // registered account state has been cleared so we do not create competing identities.
        if (auth.status === "signed_out" && !registeredAccount && !guestSession) {
          await activateGuestExperience();
        }
      };

      run()
        .catch(async (loadError: Error) => {
          if (!cancelled) {
            if (loadError instanceof ApiError && loadError.status === 401 && auth.status === "signed_in") {
              await auth.signOut();
              setError("Registered session expired. Please sign in again.");
              return;
            }

            setError(loadError.message);
          }
        })
        .finally(() => {
          if (!cancelled) {
            setIsLoading(false);
          }
        });
    });

    return () => {
      cancelled = true;
    };
  }, [auth.status, guestSession, registeredAccount]);

  useEffect(() => {
    const authJustSignedIn = previousAuthStatus.current !== "signed_in" && auth.status === "signed_in";
    previousAuthStatus.current = auth.status;

    if (auth.status === "signed_out" && registeredAccount) {
      setRegisteredAccount(null);
    }

    if (!authJustSignedIn) {
      return;
    }

    // We ask once the user signs in from a guest flow so the product can preserve or discard temporary data explicitly.
    if (guestSession && !registeredAccount) {
      setShowUpgradePrompt(true);
    }
  }, [auth.status, guestSession, registeredAccount]);

  async function refreshCoreData() {
    if (registeredAccount) {
      await loadCoreData(registeredAccount.walletID);
      return { walletID: registeredAccount.walletID, token: null };
    }

    if (guestSession) {
      try {
        await loadCoreData(guestSession.walletID, guestSession.token);
        return { walletID: guestSession.walletID, token: guestSession.token };
      } catch (refreshError) {
        if (refreshError instanceof ApiError && refreshError.status === 401) {
          const nextSession = await ensureGuestSession();
          setGuestSession(nextSession);
          await loadCoreData(nextSession.walletID, nextSession.token);
          setSuccess("Guest session refreshed.");
          return { walletID: nextSession.walletID, token: nextSession.token };
        }

        throw refreshError;
      }
    }

    return null;
  }

  async function upgradeGuestSession(preserveGuestData: boolean) {
    setIsUpgrading(true);
    setError(null);

    try {
      if (!guestSession) {
        throw new Error("Guest session is unavailable");
      }

      const registeredToken = await auth.getToken();
      if (!registeredToken) {
        throw new Error("Registered session token is unavailable");
      }

      const account = await upgradeGuestAccount({
        registeredToken,
        guestToken: guestSession.token,
        preserveGuestData
      });

      setRegisteredAccount(account);
      clearGuestSession();
      setShowUpgradePrompt(false);
      await loadCoreData(account.walletID, registeredToken);
      setSuccess(
        preserveGuestData
          ? "Guest demo data moved into the registered account."
          : "Registered demo account created with a fresh start."
      );
    } catch (upgradeError) {
      setError(upgradeError instanceof Error ? upgradeError.message : "Failed to upgrade guest account");
    } finally {
      setIsUpgrading(false);
    }
  }

  async function activeAccessToken() {
    if (registeredAccount) {
      return null;
    }

    return guestSession?.token ?? null;
  }

  function clearMessages() {
    setError(null);
    setSuccess(null);
  }

  return useMemo(
    () => ({
      guestSession,
      registeredAccount,
      markets,
      portfolio,
      orders,
      activity,
      isLoading,
      isUpgrading,
      showUpgradePrompt,
      error,
      success,
      activeWalletID,
      accountModeLabel,
      shouldShowAuthValuePrompt,
      clearMessages,
      setErrorMessage: setError,
      setSuccessMessage: setSuccess,
      refreshCoreData,
      upgradeGuestSession,
      activeAccessToken
    }),
    [
      activity,
      activeWalletID,
      accountModeLabel,
      error,
      guestSession,
      isLoading,
      isUpgrading,
      markets,
      orders,
      portfolio,
      registeredAccount,
      shouldShowAuthValuePrompt,
      showUpgradePrompt,
      success
    ]
  );
}
