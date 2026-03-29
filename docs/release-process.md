# Release Process

## Purpose

This document explains what a TradeLab release means and what gets produced when code reaches `master`.

## Release trigger

TradeLab releases are driven by the GitHub workflow chain:

`pull request -> CI -> auto-merge -> master -> release`

The release workflow runs after a pull request into `master` is merged, or manually through workflow dispatch.

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

TradeLab publishes:

- convenience tag: `latest`
- immutable release tag: `v0.1.<run-number>`

Release-ready deployment should prefer the immutable tag or the packaged manifest artifact.

## What a release signals

A release means:

- the documented CI checks passed
- build and packaging steps completed
- release artifacts were created from the merged `master` state

A release does not mean:

- the product is a live trading platform
- financial behavior is guaranteed
- all roadmap work is complete
