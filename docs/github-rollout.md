# GitHub Rollout Checklist

## Purpose

This checklist captures GitHub repository settings and public presentation details that are not fully stored as repo-tracked files.

## Repository profile

- set repository description to a product-first summary
- add relevant topics such as `go`, `nextjs`, `paper-trading`, `kubernetes`, `postgres`, `trading-sandbox`
- set homepage URL if a public product or docs URL exists

## Social preview

- use [github-social-preview.svg](assets/github-social-preview.svg) as the source asset
- export a GitHub-compatible preview image if needed
- upload it in the repository social preview settings

## Labels and planning

- create a consistent label taxonomy:
  - `type: feature`
  - `type: bug`
  - `type: documentation`
  - `type: test`
  - `type: security`
  - `type: ops`
  - `type: refactor`
  - `type: review`
  - `type: research`
  - `type: planning`
  - `area: frontend`
  - `area: backend`
  - `area: platform`
- keep exactly one `type:` label on each actively tracked open issue
- use the `v1` label on all work that is part of the TradeLab v1 release umbrella under issue `#147`
- create milestones for major product phases if public planning is desired
- optionally create a public GitHub Project with:
  - `Backlog`
  - `Planned`
  - `In Progress`
  - `Review`
  - `Released`

## Repository workflows and protection

- confirm required status checks match the documented CI jobs
- keep `squash merge` enabled
- keep branch deletion after merge enabled
- confirm branch protection applies to `master`
- enable `Allow GitHub Actions to create and approve pull requests` if the repository should let `Publish Master Images` open development-target update PRs with `GITHUB_TOKEN`
- alternatively configure a repository secret `AUTOMATION_GITHUB_TOKEN` that is allowed to create pull requests when the default Actions token is not permitted
- verify release permissions still allow image publication and GitHub Release creation

## Public repository surface

- pin or highlight the roadmap entry point if helpful
- keep README, roadmap, support, and contributing docs mutually linked
- ensure the financial-disclaimer and demo-only messaging remain visible
