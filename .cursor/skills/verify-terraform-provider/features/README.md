# Lightdash Terraform provider verification map

This directory is the maintained source for verifying practitioner-facing behavior of **terraform-provider-lightdash**. Read this index before driving Terraform CLI checks, then open the matching feature file for the exact recipe.

## Baseline preconditions

- Repository checkout at `terraform-provider-lightdash` with `.env` present (copy from `.env.template`; placeholders OK for unit tests).
- Run `.cursor/skills/verify-terraform-provider/scripts/launch.sh` once per `VERIFY_RUN_ID`.
- `terraform` on PATH matching `.terraform-version` (currently **1.14.0**).
- Local provider via dev_overrides: `ubie-oss/lightdash` → `$(go env GOPATH)/bin` (see generated `terraformrc` under `.run-state/`).
- Set `TF_CLI_CONFIG_FILE` to that `terraformrc` for all Terraform commands in a drive.
- Evidence directory: `/opt/cursor/artifacts/verify-terraform-provider/${VERIFY_RUN_ID}/`.

## Driving conventions

- Start from launch + doctor unless the feature doc says otherwise.
- Prefer **`terraform validate`** and **`terraform providers schema -json`** for offline proof; do not assume `plan` works without a real API token.
- Example HCL shapes live under `examples/data-sources/` and `examples/resources/`; scratch configs may copy those fragments into `/tmp/lightdash-tf-verify-${VERIFY_RUN_ID}/`.
- Never run `make testacc`, `TF_ACC=1`, or `integration_tests/terraform apply` as part of routine verification.
- Do not remove proof artifacts during cleanup.

## Proof and skip reporting

- CLI proof includes command, stdout/stderr, and exit code in the feature log under the evidence directory.
- Schema proof includes `provider-schema.json` or a resource snippet file, not only console text.
- If `terraform plan` is required but `LIGHTDASH_URL` / `LIGHTDASH_API_KEY` in `.env` are still placeholders, report **blocked: live API token required** with the validate/schema path you ran instead.
- Do not claim a data-source **plan/read** feature verified if you only ran validate on a different entry point.

## Feature entry contract

Each feature file uses exactly four H2 sections: **Sub-features**, **How to get to it (user POV)**, **Driving it with verify-terraform-provider**, **Gotchas**.

## Features

| ID | Doc | What it proves |
|----|-----|----------------|
| `unit-tests` | [unit-tests.md](./unit-tests.md) | Go unit tests (`make test`) without acceptance tests |
| `provider-local-install` | [provider-local-install.md](./provider-local-install.md) | Local plugin binary installed for Terraform |
| `terraform-validate-organization-ds` | [terraform-validate-organization-ds.md](./terraform-validate-organization-ds.md) | Example read-only org data source config validates offline |
| `provider-schema-json` | [provider-schema-json.md](./provider-schema-json.md) | Provider plugin schema export via Terraform CLI |
| `resource-space-schema` | [resource-space-schema.md](./resource-space-schema.md) | `lightdash_space` resource block shape in schema |
