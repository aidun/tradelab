# Contributing to TradeLab

## Purpose

TradeLab is being built as a product-grade demo trading platform. Contributions should improve the product, the engineering standards around it, or the clarity of its documentation.

## Before you contribute

Please read:

- [README.md](README.md)
- [docs/developer-guide.md](docs/developer-guide.md)
- [docs/roadmap.md](docs/roadmap.md)
- [SECURITY.md](SECURITY.md)

## Contribution standards

We expect contributions to match the repository's current quality bar:

- code and documentation remain in English
- behavior changes come with tests or an explicit documented reason they cannot
- documentation stays aligned with behavior changes
- `docs/ai-metadata.json` is updated when repo-facing structure, workflows, or public docs materially change
- CI must stay green
- public-facing claims must not overstate product maturity

## How to contribute

1. Start from the latest `master`.
2. Create a focused branch for one coherent change.
3. Keep the change reviewable and scoped.
4. Run the relevant tests locally.
5. Open a pull request using the repository template.

## Pull request expectations

Every pull request should include:

- a short summary of the change
- the reason the change matters
- issue references where applicable
- the local validation performed
- documentation updates when user-facing, operator-facing, or contributor-facing behavior changes

## Tests and validation

Use the relevant subset of:

```bash
cd backend
go test ./...
```

```bash
cd frontend
npm run test
npm run build
npm run test:e2e
```

If your change affects documentation screenshots:

```bash
cd frontend
npm run docs:screenshots
```

## Documentation expectations

TradeLab treats documentation as part of the product surface.

Update documentation when changes affect:

- user workflows
- installation or configuration
- GitHub-facing process or standards
- release behavior
- public positioning of the product

## Security reporting

Please do not open public issues for sensitive security problems. Follow the process in [SECURITY.md](SECURITY.md).

## Support boundaries

The maintainers review contributions that are aligned with the roadmap, quality bar, and product direction. Opening an issue or pull request does not guarantee merge or long-term support.
