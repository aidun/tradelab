# Onboarding Requirements

## Purpose

This document defines what the first-launch user experience must communicate and validate while TradeLab remains a guest-first, demo-only product.

## Product stance on first launch

The first launch must make these points understandable without requiring a user to read the repository:

- TradeLab is a demo trading sandbox
- the initial experience uses a guest demo session
- no live money, brokerage account, or custody relationship is involved
- market data can be fresh or temporarily stale, and the UI should communicate that clearly

## First-launch UX requirements

### Session creation

On first load, the frontend should:

1. create a guest demo session automatically
2. avoid making the user hunt for a signup flow before trying the product
3. surface an actionable error if session creation fails

### Demo-mode explanation

The UI should communicate, in visible product language, that:

- the session is a guest demo session
- balances are virtual
- trading actions are simulated
- durable account signup is optional and appears only after the user has seen product value

### Initial dashboard state

Once session bootstrap succeeds, the initial dashboard should make these surfaces visible:

- selected reference market
- chart and feed status
- wallet summary
- balances
- recent orders
- activity

### Failure handling

If first-run steps fail, the UI should prefer:

- specific error language over generic failure states
- isolated loading and error states where possible
- recovery actions such as refresh or retry guidance

## Onboarding acceptance criteria

The onboarding UX is acceptable when:

- a first-time user understands they are in demo mode
- a first-time user does not need credentials to try the product
- a first-time user can understand why and when the durable-account prompt appears
- the default market and chart state are understandable
- the first successful session-to-dashboard path is obvious
- failure states do not hide the reason or the next action

## Future compatibility

These requirements should continue to hold when registered accounts are introduced:

- guest mode remains a low-friction entry point
- registered mode becomes the durable path without replacing the demo-first experience
- the distinction between guest demo sessions and registered accounts stays explicit

## Related documentation

- [getting-started.md](getting-started.md)
- [user-guide.md](user-guide.md)
- [PRD.md](PRD.md)
