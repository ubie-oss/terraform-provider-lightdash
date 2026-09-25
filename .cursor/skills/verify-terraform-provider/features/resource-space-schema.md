# lightdash_space resource schema

Managing spaces is a core write path (`resource "lightdash_space"`). Offline verification confirms the resource type is registered and required arguments match practitioner expectations before any apply.

## Sub-features

- **validate-space-block** — Minimal resource block passes `terraform validate` (no apply).
- **schema-snippet** — Extract `lightdash_space` from `providers schema -json` into a snippet file.

## How to get to it (user POV)

- Add `resource "lightdash_space" { ... }` referencing a project UUID (see `examples/resources/lightdash_space/resource.tf`), run `terraform validate` locally with dev_overrides.

## Driving it with verify-terraform-provider

Preconditions: launch + doctor for the run.

- **Validate + snippet** — `scripts/drive.sh resource-space-schema` — Observable: validate success in log; `resource-space-schema-snippet.json` non-empty under evidence dir.

## Gotchas

- **Apply creates real spaces** in the target project; this feature never runs `apply` or `plan` with API configuration.
- Dummy `project_uuid` in the scratch config is only for validate/schema — replace with a real UUID for live plans.
- `deletion_protection` and other attributes appear in schema; behavior requires acceptance tests or manual apply (out of scope for default verification).
