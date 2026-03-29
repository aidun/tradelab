"use client";

import React, { createContext, useContext, useEffect, useMemo, useState } from "react";
import {
  ClerkProvider,
  SignInButton,
  SignUpButton,
  UserButton,
  useAuth,
  useClerk,
  useUser
} from "@clerk/nextjs";
import { logoutRegisteredAccount } from "@/lib/api";

type TradeLabAuthUser = {
  clerkUserID: string;
  email: string;
  displayName: string;
};

type TradeLabAuthContextValue = {
  available: boolean;
  provider: "none" | "mock" | "clerk";
  status: "loading" | "signed_out" | "signed_in";
  user: TradeLabAuthUser | null;
  getToken: () => Promise<string | null>;
  signInWithProvider: (provider: "google" | "apple") => void;
  signOut: () => Promise<void>;
};

const AUTH_MOCK_MODE = process.env.NEXT_PUBLIC_AUTH_MOCK_MODE === "true";
const CLERK_PUBLISHABLE_KEY = process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY;
const MOCK_STORAGE_KEY = "tradelab.mock-auth";

const TradeLabAuthContext = createContext<TradeLabAuthContextValue>({
  available: false,
  provider: "none",
  status: "signed_out",
  user: null,
  getToken: async () => null,
  signInWithProvider: () => undefined,
  signOut: async () => undefined
});

export function TradeLabAuthProvider({ children }: { children: React.ReactNode }) {
  if (AUTH_MOCK_MODE) {
    return <MockAuthProvider>{children}</MockAuthProvider>;
  }

  if (CLERK_PUBLISHABLE_KEY) {
    return (
      <ClerkProvider publishableKey={CLERK_PUBLISHABLE_KEY}>
        <ClerkStateProvider>{children}</ClerkStateProvider>
      </ClerkProvider>
    );
  }

  return <FallbackAuthProvider>{children}</FallbackAuthProvider>;
}

export function useTradeLabAuth() {
  return useContext(TradeLabAuthContext);
}

export function AuthEntryActions() {
  const auth = useTradeLabAuth();

  if (!auth.available || auth.status === "signed_in") {
    return null;
  }

  if (auth.provider === "mock") {
    return (
      <div className="flex flex-wrap gap-3">
        <button
          type="button"
          onClick={() => auth.signInWithProvider("google")}
          className="rounded-full border border-[var(--line)] px-4 py-2 text-sm font-medium text-[var(--text)] transition hover:border-[var(--accent)]"
        >
          Continue with Google
        </button>
        <button
          type="button"
          onClick={() => auth.signInWithProvider("apple")}
          className="rounded-full border border-[var(--line)] px-4 py-2 text-sm font-medium text-[var(--text)] transition hover:border-[var(--accent)]"
        >
          Continue with Apple
        </button>
      </div>
    );
  }

  if (auth.provider === "clerk") {
    return (
      <div className="flex flex-wrap gap-3">
        <SignUpButton mode="modal">
          <button
            type="button"
            className="rounded-full bg-[var(--accent)] px-4 py-2 text-sm font-semibold text-[#04111a] transition hover:brightness-105"
          >
            Sign up with Google or Apple
          </button>
        </SignUpButton>
        <SignInButton mode="modal">
          <button
            type="button"
            className="rounded-full border border-[var(--line)] px-4 py-2 text-sm font-medium text-[var(--text)] transition hover:border-[var(--accent)]"
          >
            Log in
          </button>
        </SignInButton>
      </div>
    );
  }

  return null;
}

export function AuthStatusControls() {
  const auth = useTradeLabAuth();

  if (!auth.available || auth.status !== "signed_in") {
    return null;
  }

  if (auth.provider === "clerk") {
    return <UserButton />;
  }

  return (
    <button
      type="button"
      onClick={() => {
        void auth.signOut();
      }}
      className="rounded-full border border-[var(--line)] px-4 py-2 text-sm font-medium text-[var(--text)] transition hover:border-[var(--accent)]"
    >
      Log out
    </button>
  );
}

function FallbackAuthProvider({ children }: { children: React.ReactNode }) {
  const value = useMemo<TradeLabAuthContextValue>(
    () => ({
      available: false,
      provider: "none",
      status: "signed_out",
      user: null,
      getToken: async () => null,
      signInWithProvider: () => undefined,
      signOut: async () => undefined
    }),
    []
  );

  return <TradeLabAuthContext.Provider value={value}>{children}</TradeLabAuthContext.Provider>;
}

function MockAuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<TradeLabAuthUser | null>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const raw = window.localStorage.getItem(MOCK_STORAGE_KEY);
    if (raw) {
      setUser(JSON.parse(raw) as TradeLabAuthUser);
    }
    setReady(true);
  }, []);

  const value = useMemo<TradeLabAuthContextValue>(
    () => ({
      available: true,
      provider: "mock",
      status: ready ? (user ? "signed_in" : "signed_out") : "loading",
      user,
      getToken: async () => (user ? `mock-clerk:${user.email.includes("google") ? "google" : "apple"}:${user.clerkUserID}` : null),
      signInWithProvider: (provider) => {
        const nextUser = {
          clerkUserID: provider === "google" ? "mock-google-user" : "mock-apple-user",
          email: provider === "google" ? "google-user@google.mock.tradelab" : "apple-user@apple.mock.tradelab",
          displayName: provider === "google" ? "Mock Google User" : "Mock Apple User"
        };
        window.localStorage.setItem(MOCK_STORAGE_KEY, JSON.stringify(nextUser));
        setUser(nextUser);
      },
      signOut: async () => {
        await logoutRegisteredAccount().catch(() => undefined);
        window.localStorage.removeItem(MOCK_STORAGE_KEY);
        setUser(null);
      }
    }),
    [ready, user]
  );

  return <TradeLabAuthContext.Provider value={value}>{children}</TradeLabAuthContext.Provider>;
}

function ClerkStateProvider({ children }: { children: React.ReactNode }) {
  const auth = useAuth();
  const user = useUser();
  const clerk = useClerk();

  const value = useMemo<TradeLabAuthContextValue>(
    () => ({
      available: true,
      provider: "clerk",
      status: auth.isLoaded ? (auth.isSignedIn ? "signed_in" : "signed_out") : "loading",
      user:
        auth.isSignedIn && user.user
          ? {
              clerkUserID: user.user.id,
              email: user.user.primaryEmailAddress?.emailAddress ?? "",
              displayName: user.user.fullName ?? user.user.username ?? `Trader ${user.user.id.slice(0, 8)}`
            }
          : null,
      getToken: async () => auth.getToken(),
      signInWithProvider: () => undefined,
      signOut: async () => {
        await logoutRegisteredAccount().catch(() => undefined);
        await clerk.signOut();
      }
    }),
    [auth, clerk, user.user]
  );

  return <TradeLabAuthContext.Provider value={value}>{children}</TradeLabAuthContext.Provider>;
}
