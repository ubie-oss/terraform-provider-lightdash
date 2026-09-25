#!/usr/bin/env bash
# Drive one mapped feature; write proof artifacts to VERIFY_EVIDENCE_DIR.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "${SCRIPT_DIR}/common.sh"

FEATURE="${1-}"
if [[ -z ${FEATURE} ]]; then
	echo "Usage: $0 <feature-id>" >&2
	echo "See features/README.md for ids: unit-tests, provider-local-install, terraform-validate-organization-ds, provider-schema-json, resource-space-schema" >&2
	exit 2
fi

export TF_CLI_CONFIG_FILE="${TERRAFORMRC}"

record_meta() {
	local recorded_at
	recorded_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	{
		echo "feature=${FEATURE}"
		echo "run_id=${VERIFY_RUN_ID}"
		echo "recorded_at=${recorded_at}"
	} >>"${VERIFY_EVIDENCE_DIR}/run.meta"
}

run_unit_tests() {
	ensure_env_file
	local log="${VERIFY_EVIDENCE_DIR}/unit-tests.log"
	echo "==> make test (TF_ACC unset)" | tee "${log}"
	(cd "${REPO_ROOT}" && make test) 2>&1 | tee -a "${log}"
	local ec="${PIPESTATUS[0]}"
	echo "exit_code=${ec}" >>"${VERIFY_EVIDENCE_DIR}/unit-tests.meta"
	[[ ${ec} -eq 0 ]]
}

run_provider_local_install() {
	local log="${VERIFY_EVIDENCE_DIR}/provider-install.log"
	{
		echo "==> go install ."
		(cd "${REPO_ROOT}" && go install -v .)
		echo "==> binary inspect"
		ls -la "${GOBIN}/terraform-provider-lightdash"
		"${GOBIN}/terraform-provider-lightdash" 2>&1 || true
	} | tee "${log}"
	# Provider exits non-zero when run directly without terraform; presence is the proof.
	test -x "${GOBIN}/terraform-provider-lightdash"
}

write_validate_workspace() {
	local dir="${VERIFY_SCRATCH_ROOT}/validate-organization-ds"
	mkdir -p "${dir}"
	cat >"${dir}/main.tf" <<'EOF'
terraform {
  required_version = ">= 1.14.0"
  required_providers {
    lightdash = {
      source = "ubie-oss/lightdash"
    }
  }
}

provider "lightdash" {
  host  = "https://example.invalid"
  token = "offline-validate-only-not-a-real-token"
}

data "lightdash_organization" "example" {}
EOF
	echo "${dir}"
}

run_terraform_validate_organization_ds() {
	local dir
	dir="$(write_validate_workspace)"
	local log="${VERIFY_EVIDENCE_DIR}/terraform-validate-organization-ds.log"
	{
		echo "==> workspace ${dir}"
		echo "==> terraform validate (dev_overrides via TF_CLI_CONFIG_FILE=${TERRAFORMRC})"
		(cd "${dir}" && terraform validate -no-color)
	} 2>&1 | tee "${log}"
}

run_provider_schema_json() {
	local dir
	dir="$(write_validate_workspace)"
	local out="${VERIFY_EVIDENCE_DIR}/provider-schema.json"
	local log="${VERIFY_EVIDENCE_DIR}/provider-schema-json.log"
	{
		echo "==> terraform validate in ${dir}"
		(cd "${dir}" && terraform validate -no-color)
		echo "==> terraform providers schema -json"
		(cd "${dir}" && terraform providers schema -json) >"${out}"
		wc -c "${out}"
		head -c 400 "${out}"
		echo ""
	} 2>&1 | tee "${log}"
	test -s "${out}"
}

run_resource_space_schema() {
	local dir="${VERIFY_SCRATCH_ROOT}/schema-space"
	mkdir -p "${dir}"
	cat >"${dir}/main.tf" <<'EOF'
terraform {
  required_providers {
    lightdash = { source = "ubie-oss/lightdash" }
  }
}
provider "lightdash" {
  host  = "https://example.invalid"
  token = "offline-validate-only-not-a-real-token"
}
resource "lightdash_space" "example" {
  name        = "verify-scratch"
  project_uuid = "00000000-0000-0000-0000-000000000000"
}
EOF
	local log="${VERIFY_EVIDENCE_DIR}/resource-space-schema.log"
	local snippet="${VERIFY_EVIDENCE_DIR}/resource-space-schema-snippet.json"
	{
		echo "==> terraform validate (schema only, no apply)"
		(cd "${dir}" && terraform validate -no-color)
		echo "==> extract lightdash_space from providers schema"
		(cd "${dir}" && terraform providers schema -json) |
			python3 -c "import json,sys; s=json.load(sys.stdin); r=s['provider_schemas']['registry.terraform.io/ubie-oss/lightdash']['resource_schemas'].get('lightdash_space'); print(json.dumps({'lightdash_space': r}, indent=2)[:8000])" \
				>"${snippet}"
		wc -c "${snippet}"
	} 2>&1 | tee "${log}"
	test -s "${snippet}"
}

record_meta

case "${FEATURE}" in
unit-tests)
	run_unit_tests
	;;
provider-local-install)
	run_provider_local_install
	;;
terraform-validate-organization-ds)
	run_terraform_validate_organization_ds
	;;
provider-schema-json)
	run_provider_schema_json
	;;
resource-space-schema)
	run_resource_space_schema
	;;
*)
	echo "Unknown feature: ${FEATURE}" >&2
	exit 2
	;;
esac

echo "Drive OK: ${FEATURE} — evidence in ${VERIFY_EVIDENCE_DIR}"
