"use client";

import { startTransition, useEffect, useMemo, useRef, useState } from "react";

import {
  type AccountingMode,
  ApiError,
  bootstrapRegisteredAccount,
  createDemoSession,
  fetchActivity,
  fetchMarkets,
  fetchOrders,
  fetchPortfolio,
  fetchStrategies,
  upgradeGuestAccount,
  type ActivityLog,
  type DemoSession,
  type Market,
  type Order,
  type PortfolioSummary,
  type RegisteredAccount,
  type Strategy
} from "@/lib/api";
import {
  clearStoredDemoSession,
  readStoredDemoSession,
  storeDemoSession
} from "@/lib/demo-session-storage";
import { useTradeLabAuth } from "@/lib/tradelab-auth";
const ACCOUNTING_MODE_STORAGE_KEY = "tradelab.accounting-mode";

type CoreDataState = {
  guestSession: DemoSession | null;
  registeredAccount: RegisteredAccount | null;
  markets: Market[];
  portfolio: PortfolioSummary | null;
  orders: Order[];
  activity: ActivityLog[];
  activityError: string | null;
  strategies: Strategy[];
  accountingMode: AccountingMode;
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
  setAccountingMode: (mode: AccountingMode) => void;
  startGuestSession: () => Promise<{ walletID: string; token: string | null }>;
  refreshCoreData: () => Promise<{ walletID: string; token: string | null } | null>;
  upgradeGuestSession: (preserveGuestData: boolean) => Promise<void>;
  activeAccessToken: () => Promise<string | null>;
};

type UseAccountSessionOptions = {
  autoStartGuest?: boolean;
};

/** useAccountSession orchestrates guest and registered account state plus dashboard data loading. */
export function useAccountSession(options?: UseAccountSessionOptions): CoreDataState {
  const auth = useTradeLabAuth();
  const previousAuthStatus = useRef(auth.status);
  const hasHydratedAccountingMode = useRef(false);
  const autoStartGuest = options?.autoStartGuest ?? true;

  const [guestSession, setGuestSession] = useState<DemoSession | null>(null);
  const [registeredAccount, setRegisteredAccount] = useState<RegisteredAccount | null>(null);
  const [markets, setMarkets] = useState<Market[]>([]);
  const [portfolio, setPortfolio] = useState<PortfolioSummary | null>(null);
  const [orders, setOrders] = useState<Order[]>([]);
  const [activity, setActivity] = useState<ActivityLog[]>([]);
  const [activityError, setActivityError] = useState<string | null>(null);
  const [strategies, setStrategies] = useState<Strategy[]>([]);
  const [accountingMode, setAccountingModeState] = useState<AccountingMode>("average_cost");
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

    const cachedSession = readStoredDemoSession();
    if (cachedSession) {
      return cachedSession;
    }

    const nextSession = await createDemoSession();
    storeDemoSession(nextSession);
    return nextSession;
  }

  function clearGuestSession() {
    clearStoredDemoSession();
    setGuestSession(null);
  }

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    const storedMode = window.localStorage.getItem(ACCOUNTING_MODE_STORAGE_KEY) as AccountingMode | null;
    if (storedMode === "average_cost" || storedMode === "fifo" || storedMode === "hybrid") {
      setAccountingModeState(storedMode);
    }
  }, []);

  function resolveErrorMessage(reason: unknown, fallback: string) {
    if (reason instanceof Error && reason.message.trim() !== "") {
      return reason.message;
    }

    return fallback;
  }

  async function loadCoreData(walletID: string, token?: string | null) {
    const results = await Promise.allSettled([
      fetchMarkets(),
      fetchPortfolio(walletID, token ?? "", accountingMode),
      fetchOrders(token ?? "", { accountingMode }),
      fetchActivity(token ?? ""),
      fetchStrategies(token ?? "")
    ]);
    const [marketList, portfolioSummary, orderHistory, activityHistory, strategyList] = results;

    const authFailure = results.find(
      (result): result is PromiseRejectedResult =>
        result.status === "rejected" && result.reason instanceof ApiError && result.reason.status === 401
    );
    if (authFailure) {
      throw authFailure.reason;
    }

    let nextError: string | null = null;

    if (marketList.status === "fulfilled") {
      setMarkets(marketList.value);
    } else {
      nextError ??= resolveErrorMessage(marketList.reason, "Failed to load markets");
    }

    if (portfolioSummary.status === "fulfilled") {
      setPortfolio(portfolioSummary.value);
    } else {
      nextError ??= resolveErrorMessage(portfolioSummary.reason, "Failed to load portfolio");
    }

    if (orderHistory.status === "fulfilled") {
      setOrders(orderHistory.value);
    } else {
      nextError ??= resolveErrorMessage(orderHistory.reason, "Failed to load orders");
    }

    if (activityHistory.status === "fulfilled") {
      setActivity(activityHistory.value);
      setActivityError(null);
    } else {
      setActivityError(resolveErrorMessage(activityHistory.reason, "Failed to load activity"));
    }

    if (strategyList.status === "fulfilled") {
      setStrategies(strategyList.value);
    } else {
      nextError ??= resolveErrorMessage(strategyList.reason, "Failed to load strategies");
    }

    setError(nextError);
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
        if (autoStartGuest && auth.status === "signed_out" && !registeredAccount && !guestSession) {
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
  }, [auth.status, autoStartGuest, guestSession, registeredAccount]);

  useEffect(() => {
    if (!activeWalletID) {
      return;
    }

    if (!hasHydratedAccountingMode.current) {
      hasHydratedAccountingMode.current = true;
      return;
    }

    startTransition(() => {
      refreshCoreData().catch((refreshError: Error) => {
        setError(refreshError.message);
      });
    });
  }, [accountingMode]);

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

  async function startGuestSession() {
    setError(null);
    setSuccess(null);
    setIsLoading(true);

    try {
      return await activateGuestExperience();
    } finally {
      setIsLoading(false);
    }
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

  function setAccountingMode(mode: AccountingMode) {
    setAccountingModeState(mode);
    if (typeof window !== "undefined") {
      window.localStorage.setItem(ACCOUNTING_MODE_STORAGE_KEY, mode);
    }
  }

  return useMemo(
    () => ({
      guestSession,
      registeredAccount,
      markets,
      portfolio,
      orders,
      activity,
      activityError,
      strategies,
      accountingMode,
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
      setAccountingMode,
      startGuestSession,
      refreshCoreData,
      upgradeGuestSession,
      activeAccessToken
    }),
    [
      activity,
      activityError,
      accountingMode,
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
      startGuestSession,
      strategies,
      shouldShowAuthValuePrompt,
      showUpgradePrompt,
      success
    ]
  );
}
