#!/usr/bin/env bash
# Remove scratch workspaces and run-state; never delete evidence under VERIFY_EVIDENCE_DIR.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "${SCRIPT_DIR}/common.sh"

if [[ -d "${VERIFY_SCRATCH_ROOT}" ]]; then
  rm -rf "${VERIFY_SCRATCH_ROOT}"
  echo "Removed scratch: ${VERIFY_SCRATCH_ROOT}"
fi

if [[ -d "${VERIFY_STATE_DIR}" ]]; then
  rm -rf "${VERIFY_STATE_DIR}"
  echo "Removed run-state: ${VERIFY_STATE_DIR}"
fi

if [[ -d "${VERIFY_EVIDENCE_DIR}" ]]; then
  echo "Evidence preserved: ${VERIFY_EVIDENCE_DIR}"
else
  echo "No evidence directory at ${VERIFY_EVIDENCE_DIR}"
fi
