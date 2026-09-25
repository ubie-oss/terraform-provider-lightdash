#!/usr/bin/env bash
# Build/install the provider and write a disposable dev_overrides CLI config.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "${SCRIPT_DIR}/common.sh"

ensure_env_file
ensure_terraform

echo "==> Installing provider from ${REPO_ROOT}"
(cd "${REPO_ROOT}" && go install -v .)

if [[ ! -x "${GOBIN}/terraform-provider-lightdash" ]]; then
	echo "Expected binary missing: ${GOBIN}/terraform-provider-lightdash" >&2
	exit 1
fi

mkdir -p "${VERIFY_STATE_DIR}"
cat >"${TERRAFORMRC}" <<EOF
provider_installation {
  dev_overrides {
    "${PROVIDER_SOURCE}" = "${GOBIN}"
  }
  direct {}
}
EOF

launched_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
{
	echo "run_id=${VERIFY_RUN_ID}"
	echo "gobin=${GOBIN}"
	echo "terraformrc=${TERRAFORMRC}"
	echo "provider_binary=${GOBIN}/terraform-provider-lightdash"
	echo "launched_at=${launched_at}"
} >"${VERIFY_STATE_DIR}/launch.meta"

echo "Launch OK: provider installed, dev_overrides at ${TERRAFORMRC}"
echo "Evidence directory: ${VERIFY_EVIDENCE_DIR}"
