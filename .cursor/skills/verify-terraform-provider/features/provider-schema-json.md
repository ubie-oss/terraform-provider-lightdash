# Provider schema JSON export

Practitioners and tooling inspect registered data sources, resources, and provider attributes via `terraform providers schema -json`. This feature exports that document from the **locally overridden** plugin after validate.

## Sub-features

- **schema-json-file** — Full JSON written to `provider-schema.json` in the evidence directory.
- **provider-block-attributes** — JSON includes `host`, `token`, optional `max_concurrent_requests`.

## How to get to it (user POV)

- In a Terraform workspace with the Lightdash provider configured, run `terraform providers schema -json` (often after `terraform validate`).

## Driving it with verify-terraform-provider

Preconditions: same as organization validate feature (launch + doctor).

- **Export schema** — `scripts/drive.sh provider-schema-json` — Observable: `provider-schema.json` exists, non-empty; log records byte count and a short prefix of JSON.

## Gotchas

- Schema reflects **local binary** when dev_overrides are active; version string may differ from registry release.
- Large JSON; proof uses file size and snippet, not manual diff of entire schema each run.
- Does not prove individual CRUD operations against Lightdash API.
