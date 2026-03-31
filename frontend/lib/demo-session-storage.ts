"use client";

import type { DemoSession } from "@/lib/api";

export const DEMO_SESSION_STORAGE_KEY = "tradelab.demo-session";

/** readStoredDemoSession restores a still-valid guest session from session storage. */
export function readStoredDemoSession(): DemoSession | null {
  if (typeof window === "undefined") {
    return null;
  }

  const cachedValue = window.sessionStorage.getItem(DEMO_SESSION_STORAGE_KEY);
  if (!cachedValue) {
    return null;
  }

  try {
    const cachedSession = JSON.parse(cachedValue) as DemoSession;
    if (new Date(cachedSession.expiresAt).getTime() > Date.now()) {
      return cachedSession;
    }
  } catch {
    // Ignore malformed session payloads and clear them below.
  }

  window.sessionStorage.removeItem(DEMO_SESSION_STORAGE_KEY);
  return null;
}

/** storeDemoSession persists the active guest session for reuse within the browser session. */
export function storeDemoSession(session: DemoSession) {
  if (typeof window === "undefined") {
    return;
  }

  window.sessionStorage.setItem(DEMO_SESSION_STORAGE_KEY, JSON.stringify(session));
}

/** clearStoredDemoSession removes the guest session from browser session storage. */
export function clearStoredDemoSession() {
  if (typeof window === "undefined") {
    return;
  }

  window.sessionStorage.removeItem(DEMO_SESSION_STORAGE_KEY);
}
