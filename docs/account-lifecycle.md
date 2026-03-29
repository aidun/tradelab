# Account Lifecycle

## Purpose

This document defines how guest sessions and registered accounts should behave over time in TradeLab.

## Lifecycle types

TradeLab should support two lifecycle types:

- guest lifecycle
- registered account lifecycle

## Guest lifecycle

### Creation

- created automatically on first launch
- attached to a temporary demo wallet

### Active use

- user can inspect markets and place demo trades
- state is temporary and should not be treated as durable ownership

### Expiry

- guest sessions should expire automatically after their time window
- expired guest sessions should not pretend to preserve long-term continuity

### Upgrade

Guest users should be able to upgrade to a registered account.

The upgrade path now lets the user choose whether:

- demo balances are preserved
- trade history is preserved
- a new durable account is created and linked

## Registered account lifecycle

### Creation

- created through Clerk signup or first successful social login
- mapped to an internal TradeLab user record

### Active use

- supports repeat access across sessions and devices
- owns durable demo-account state

### Session loss or expiry

- identity re-authentication happens through Clerk
- durable product data remains attached to the registered account

### Logout

- user leaves the durable account context
- the product returns to guest re-entry automatically

## Multi-device expectations

Registered accounts should support multi-device continuity.

This means:

- the same user should regain the same durable demo state after login
- device switching should not fork identity unintentionally

## Retention and cleanup expectations

TradeLab should eventually define:

- guest-session cleanup rules
- registered-account retention expectations
- how long expired guest data is kept, if at all

These retention rules should later align with the Phase 3 security and sensitive-data handling work.

## Related documentation

- [authentication-model.md](authentication-model.md)
- [auth-flows.md](auth-flows.md)
- [clerk-architecture.md](clerk-architecture.md)
