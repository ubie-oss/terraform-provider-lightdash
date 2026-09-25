#!/usr/bin/env bash
# Read-only preflight: toolchain, repo .env, local provider binary, dev_overrides file.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "${SCRIPT_DIR}/common.sh"

fail=0
report() { echo "$*"; }

report "==> verify-terraform-provider doctor (run_id=${VERIFY_RUN_ID})"

if [[ ! -f "${REPO_ROOT}/.env" ]]; then
	report "WARN: ${REPO_ROOT}/.env missing (copy from .env.template before make test)"
	fail=1
else
	report "OK: .env present"
fi

if ! command -v go >/dev/null 2>&1; then
	report "FAIL: go not on PATH"
	fail=1
else
	go_version="$(go version)"
	report "OK: ${go_version}"
fi

set +e
ensure_terraform
terraform_check=$?
set -e
if [[ ${terraform_check} -eq 0 ]]; then
	tf_version_line="$(terraform version | head -1)"
	report "OK: ${tf_version_line}"
else
	fail=1
fi

if [[ ! -x "${GOBIN}/terraform-provider-lightdash" ]]; then
	report "FAIL: run scripts/launch.sh first (${GOBIN}/terraform-provider-lightdash missing)"
	fail=1
else
	report "OK: provider binary ${GOBIN}/terraform-provider-lightdash"
fi

if [[ ! -f ${TERRAFORMRC} ]]; then
	report "FAIL: missing ${TERRAFORMRC} (run scripts/launch.sh)"
	fail=1
else
	if grep -q "dev_overrides" "${TERRAFORMRC}" && grep -q "${PROVIDER_SOURCE}" "${TERRAFORMRC}"; then
		report "OK: dev_overrides configured for ${PROVIDER_SOURCE}"
	else
		report "FAIL: ${TERRAFORMRC} does not reference ${PROVIDER_SOURCE}"
		fail=1
	fi
fi

if [[ ${fail} -ne 0 ]]; then
	report "Doctor: NOT READY"
	exit 1
fi

report "Doctor: READY"
exit 0
