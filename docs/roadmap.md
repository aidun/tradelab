# Public Roadmap

## Purpose

This roadmap gives a high-level view of where TradeLab is headed as a product. It is intentionally curated: it highlights direction and priorities rather than every internal task.

## Current product themes

### 1. Trading simulation quality

Focus:

- more realistic execution behavior
- clearer portfolio valuation
- better activity and history visibility
- stronger trust in demo outcomes

### 2. Strategy and automation

Focus:

- first strategy engine capabilities
- configurable rule-based automation
- backtesting foundations
- clearer explanation of automated decisions

### 3. Platform and observability

Focus:

- stronger logging and monitoring posture
- better release confidence
- cleaner operational playbooks
- improved deployment maturity

### 4. Product UX and decision support

Focus:

- richer market detail flows
- clearer chart and portfolio storytelling
- stronger visual polish
- more usable onboarding and documentation

### 5. Identity and trust

Focus:

- move from guest-only access to a dual model of guest sessions plus durable registered accounts
- introduce Clerk as the managed auth provider
- support Google and Apple sign-in
- define account lifecycle, session behavior, and upgrade paths before live-like expansion

## Near-term priorities

- keep the trading sandbox stable and understandable
- improve public product and engineering documentation
- define the authentication and account roadmap before implementing registered access
- strengthen roadmap, release, and GitHub contribution signals
- continue increasing quality gates around CI, E2E coverage, and release automation

## How we plan work

TradeLab currently follows this planning and delivery model:

1. work is scoped into focused feature branches
2. pull requests are validated by CI
3. successful pull requests are squash-merged into `master`
4. `master` triggers release verification and artifact publication

## What this roadmap is not

This roadmap is not:

- a commitment that every item will ship on a fixed date
- a guarantee of feature support timelines
- a replacement for detailed implementation issues or release notes
