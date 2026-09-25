#!/usr/bin/env bash
# End-to-end wrapper: launch → doctor → drive one feature → cleanup; evidence must survive.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FEATURE="${1:-terraform-validate-organization-ds}"

export VERIFY_RUN_ID="${VERIFY_RUN_ID:-verify-$(date -u +%Y%m%dT%H%M%SZ)-$$}"
# shellcheck source=common.sh
source "${SCRIPT_DIR}/common.sh"

"${SCRIPT_DIR}/launch.sh"
"${SCRIPT_DIR}/doctor.sh"
"${SCRIPT_DIR}/drive.sh" "${FEATURE}"
"${SCRIPT_DIR}/cleanup.sh"

evidence_entries="$(ls -A "${VERIFY_EVIDENCE_DIR}" 2>/dev/null || true)"
if [[ ! -d ${VERIFY_EVIDENCE_DIR} ]] || [[ -z ${evidence_entries} ]]; then
	echo "FAIL: evidence missing after cleanup" >&2
	exit 1
fi
echo "Proof complete: evidence survived cleanup at ${VERIFY_EVIDENCE_DIR}"
ls -la "${VERIFY_EVIDENCE_DIR}"
