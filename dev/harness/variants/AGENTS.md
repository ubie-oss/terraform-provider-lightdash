# AGENTS.md

Shared norms for coding agents in this repo. How each tool loads this file is recorded in `dev/harness/agent-discovery.md`, which is not injected automatically.

## Overview

Terraform provider for Lightdash, written in Go. One module. No local services; the provider binary talks to a remote Lightdash API.

## Commands

Defined in `GNUmakefile`.

- Unit tests: `make test` runs `go test` inside `internal/` and skips acceptance tests. `GNUmakefile` includes `.env` unconditionally (line 19), so copy `.env.template` to `.env` before any `make` target. Placeholder values are enough for unit tests. Acceptance tests need real credentials.
- Compile: `go build -v ./`. `make build` also runs doc generation, `go mod tidy`, `gosec`, and `deadcode`. `go generate ./...` can fail with a template error on `lightdash_authenticated_user`. That failure is pre-existing and does not affect the compile or unit tests.
- Lint: `trunk check --all`. `make lint` also runs `pre-commit run --all-files`.
- Format: `make format` runs `go fmt` and `trunk fmt --all`.
- Acceptance: `make testacc` against a live Lightdash instance. See `CONTRIBUTING.md`.

`gosec` is installed under `~/go/bin`. `pre-commit` is installed under `~/.local/bin`.

## Local provider binary

1. `go install .`
2. In `~/.terraformrc`, set a `dev_overrides` entry for `ubie-oss/lightdash` to `$GOPATH/bin`.
3. `terraform plan` in a directory with a Lightdash provider config loads that binary. Plans and applies need a real API token.

More setup is in `CONTRIBUTING.md`.

## Constraints

- When implementing an attached plan, leave the plan file unchanged and use the existing todos.
- High-impact credentials (for example OAuth clients) use `deletion_protection` the same way `lightdash_space` does: a Terraform-only delete guard, default `true` on import.
- Do not add acceptance tests for `lightdash_project_role_member` or `lightdash_organization_role_member`. They change real user org and project access.
- Unit and integration tests stay off the live API when a call would change org or project access. Use JSON fixtures and pure service logic.
- Shared Terraform string Set/List conversion helpers belong in `internal/provider/utils.go`.
- Role IAM behavior, embedded docs under `internal/provider/docs/`, and the `tfplugindocs` `--provider-name` quirk are in `.cursor/rules/role-iam.mdc` and `.cursor/rules/provider-docs.mdc`. They attach when the matching files are in context.
