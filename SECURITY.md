# Security Policy

## Scope

TradeLab is a demo-only trading sandbox, but it still handles authentication, persistence, deployment assets, and release automation. Security issues are treated seriously.

## Supported surface

Security reports are relevant for:

- backend authentication and authorization
- dependency or container vulnerabilities
- deployment and secret handling
- CI/CD and release automation
- any bug that could expose data, credentials, or unauthorized behavior

## How to report a vulnerability

Please do not report security vulnerabilities through public GitHub issues.

Use a private channel instead:

- GitHub Security Advisories, if enabled for the repository
- or direct maintainer contact through the support path documented in [SUPPORT.md](SUPPORT.md)

Your report should include:

- a clear description of the issue
- affected area and impact
- reproduction steps if safe to share
- mitigation ideas if you have them

## Response expectations

The maintainers aim to:

- acknowledge the report quickly
- validate severity and affected scope
- prepare a fix and release path
- avoid exposing sensitive implementation details publicly before remediation

## Responsible disclosure

Please avoid public disclosure until the issue has been reviewed and a remediation path is available.

## Out of scope

The following are generally out of scope unless they create a concrete exploit path:

- issues that depend on unsupported local configurations
- theoretical attacks without a realistic impact path
- problems caused only by obviously insecure operator choices already documented as unsafe
