# Release Process

## Purpose

This document explains what a TradeLab release means, how it is triggered, and how production is promoted afterward.

## Release trigger

TradeLab releases are driven by the GitHub workflow chain:

`pull request -> CI -> auto-merge -> master -> publish master images for development -> manual release workflow`

The release workflow is intentionally manual and is triggered through GitHub Actions `workflow_dispatch`.

Recommended operator flow:

1. merge reviewed work into `master`
2. trigger the `Release` workflow from `master`
3. confirm the GitHub Release and immutable images were published
4. trigger `Promote Production` when production should move to the newest official release

## Release stages

1. release metadata is generated
2. backend and frontend are verified again
3. backend binaries and frontend artifacts are built
4. backend and frontend images are published to GHCR
5. immutable Kubernetes manifests are rendered and packaged
6. a GitHub Release is created with the generated artifacts

## Published artifacts

Each successful release currently publishes:

- Linux backend binary
- Windows backend binary
- frontend standalone artifact
- Kubernetes manifest package
- backend container image
- frontend container image

## Image tagging

TradeLab publishes two image classes:

- development tags from merged `master` commits:
  - `master`
  - `master-<shortsha>`
- official release tags:
  - `latest`
  - `v0.1.<run-number>`

`tradelab-dev` should use immutable `master-<shortsha>` image tags through its Argo CD application manifest. `tradelab-prod` should use immutable official release tags or the packaged manifest artifact.

## Environment policy

- `tradelab-dev` always follows `master`
- `tradelab-prod` is promoted only to an official GitHub release tag
- production promotion is handled by the `Promote Production` workflow, which updates the Argo CD production application to the selected or latest published release

## What a release signals

A release means:

- the documented CI checks passed
- build and packaging steps completed
- release artifacts were created from the selected `master` ref and published as an official immutable release

A release does not mean:

- the product is a live trading platform
- financial behavior is guaranteed
- all roadmap work is complete
