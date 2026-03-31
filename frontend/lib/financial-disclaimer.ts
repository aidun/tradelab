"use client";

export type FinancialDisclaimerIdentity =
  | { type: "guest"; walletID: string }
  | { type: "account"; accountID: string };

function storageKey(identity: FinancialDisclaimerIdentity) {
  return identity.type === "guest"
    ? `tradelab.disclaimer.guest.${identity.walletID}`
    : `tradelab.disclaimer.account.${identity.accountID}`;
}

function readValue(identity: FinancialDisclaimerIdentity) {
  if (typeof window === "undefined") {
    return null;
  }

  const key = storageKey(identity);
  return identity.type === "guest"
    ? window.sessionStorage.getItem(key)
    : window.localStorage.getItem(key);
}

/** hasAcknowledgedFinancialDisclaimer returns whether the current identity already acknowledged the v1 disclaimer. */
export function hasAcknowledgedFinancialDisclaimer(identity: FinancialDisclaimerIdentity) {
  return readValue(identity) === "acknowledged";
}

/** acknowledgeFinancialDisclaimer stores the one-time disclaimer acknowledgement for the current identity. */
export function acknowledgeFinancialDisclaimer(identity: FinancialDisclaimerIdentity) {
  if (typeof window === "undefined") {
    return;
  }

  const key = storageKey(identity);
  if (identity.type === "guest") {
    window.sessionStorage.setItem(key, "acknowledged");
    return;
  }

  window.localStorage.setItem(key, "acknowledged");
}
