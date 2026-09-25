# Validate organization data source (offline)

Reading organization metadata is a common first Terraform step (`data "lightdash_organization"`). Offline verification ensures example-shaped HCL validates against the **local** provider without calling the API.

## Sub-features

- **validate-org-ds** — Scratch config mirroring `examples/data-sources/lightdash_organization/data-source.tf` plus provider block validates successfully.
- **dev-overrides-active** — Terraform warns that provider development overrides are in effect.

## How to get to it (user POV)

- Author Terraform with `provider "lightdash" { host, token }` and `data "lightdash_organization" "example" {}`, then run `terraform validate` in a workspace with local dev_overrides (see `README.md` usage guide).

## Driving it with verify-terraform-provider

Preconditions: `launch.sh` + `doctor.sh` for the run; `TF_CLI_CONFIG_FILE` points at generated `terraformrc`.

- **Validate workspace** — `scripts/drive.sh terraform-validate-organization-ds` — Observable: log contains `Success! The configuration is valid`; exit code `0`.

## Gotchas

- **`terraform validate` does not verify API connectivity** or token validity; `plan`/`apply` will call `GetMyOrganizationV1` and fail with invalid tokens.
- HashiCorp recommends **skipping `terraform init`** when dev_overrides are set; this drive uses validate-only without init.
- Live **plan** for this data source requires real `host` + token in the provider block or variables.
