#!/usr/bin/env bash
# Shared paths and environment for verify-terraform-provider.
set -euo pipefail

REPO_ROOT="${REPO_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)}"
export REPO_ROOT

export PATH="${HOME}/.local/bin:${PATH}:$(go env GOPATH 2>/dev/null)/bin"

VERIFY_RUN_ID="${VERIFY_RUN_ID:-verify-$(date -u +%Y%m%dT%H%M%SZ)-$$}"
export VERIFY_RUN_ID

VERIFY_EVIDENCE_DIR="${VERIFY_EVIDENCE_DIR:-/opt/cursor/artifacts/verify-terraform-provider/${VERIFY_RUN_ID}}"
export VERIFY_EVIDENCE_DIR

VERIFY_SCRATCH_ROOT="${VERIFY_SCRATCH_ROOT:-/tmp/lightdash-tf-verify-${VERIFY_RUN_ID}}"
export VERIFY_SCRATCH_ROOT

VERIFY_STATE_DIR="${VERIFY_STATE_DIR:-${REPO_ROOT}/.cursor/skills/verify-terraform-provider/.run-state/${VERIFY_RUN_ID}}"
export VERIFY_STATE_DIR

PROVIDER_SOURCE="${PROVIDER_SOURCE:-ubie-oss/lightdash}"
export PROVIDER_SOURCE

GOBIN="$(go env GOPATH)/bin"
export GOBIN

TERRAFORMRC="${TERRAFORMRC:-${VERIFY_STATE_DIR}/terraformrc}"
export TERRAFORMRC

ensure_env_file() {
  if [[ ! -f "${REPO_ROOT}/.env" ]]; then
    cp "${REPO_ROOT}/.env.template" "${REPO_ROOT}/.env"
  fi
}

ensure_terraform() {
  if ! command -v terraform >/dev/null 2>&1; then
    echo "terraform not found on PATH; install Terraform $(cat "${REPO_ROOT}/.terraform-version" 2>/dev/null || echo 1.14.0) (see CONTRIBUTING.md)." >&2
    return 1
  fi
}

mkdir -p "${VERIFY_EVIDENCE_DIR}" "${VERIFY_STATE_DIR}"
