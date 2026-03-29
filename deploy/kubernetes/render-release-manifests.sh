#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <release-tag> <output-file>" >&2
  exit 1
fi

RELEASE_TAG="$1"
OUTPUT_FILE="$2"

kubectl kustomize deploy/kubernetes/overlays/production \
  | sed "s/IMAGE_TAG_PLACEHOLDER/${RELEASE_TAG}/g" \
  > "${OUTPUT_FILE}"

if grep -q "IMAGE_TAG_PLACEHOLDER" "${OUTPUT_FILE}"; then
  echo "placeholder tag still present in rendered manifests" >&2
  exit 1
fi

if grep -q "ghcr.io/aidun/tradelab-.*:latest" "${OUTPUT_FILE}"; then
  echo "mutable latest tag found in rendered release manifests" >&2
  exit 1
fi
