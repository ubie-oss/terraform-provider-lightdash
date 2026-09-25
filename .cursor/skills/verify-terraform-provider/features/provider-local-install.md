# Local provider install

Terraform practitioners override the registry provider with a locally built binary when developing the plugin. This feature proves `go install` produced `terraform-provider-lightdash` in `GOPATH/bin`.

## Sub-features

- **go-install** — Build and install from repo root (`go install .`).
- **binary-on-path** — Executable present at `$(go env GOPATH)/bin/terraform-provider-lightdash`.

## How to get to it (user POV)

- Follow `CONTRIBUTING.md` / `AGENTS.md`: `go install .` and configure `dev_overrides` in `~/.terraformrc` (this skill uses a disposable `terraformrc` instead).

## Driving it with verify-terraform-provider

Preconditions: `scripts/launch.sh` completed for the same `VERIFY_RUN_ID`.

- **Install and inspect** — `scripts/drive.sh provider-local-install` — Observable: log lists the binary; file is executable; size non-zero.

## Gotchas

- Registry address in code is `registry.terraform.io/ubie-oss/terraform-provider-lightdash`; **`required_providers` source is `ubie-oss/lightdash`** — dev_overrides must use the latter.
- Running the provider binary directly without Terraform exits immediately; proof is file presence, not a daemon.
- `CONTRIBUTING.md` shows a legacy override key `github.com/ubie-oss/terraform-provider-lightdash`; prefer **`ubie-oss/lightdash`** to match `README.md`.
