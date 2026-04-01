# AGENTS.md

## Cursor Cloud specific instructions

### Overview

This is a Terraform provider for Lightdash, written in Go. It is a single-product repo (not a monorepo). There are no local services to run — the provider binary communicates with a remote Lightdash API instance.

### Key commands

All standard dev commands are defined in `GNUmakefile`:

- **Unit tests:** `make test` (runs `go test` inside `internal/`, skipping acceptance tests)
- **Build:** `go build -v ./` (the plain Go build; `make build` also runs `gen-docs`, `go-tidy`, `gosec`, `deadcode`)
- **Lint:** `trunk check --all` (the `make lint` target also runs `pre-commit run --all-files`)
- **Format:** `make format` (runs `go fmt` + `trunk fmt --all`)
- **Acceptance tests:** `make testacc` (requires a live Lightdash instance; see `CONTRIBUTING.md`)

### Gotchas

- **`.env` file is required.** The `GNUmakefile` has `include .env` at line 19, which is unconditional. If `.env` does not exist, every `make` target fails — even `make test`. Copy `.env.template` to `.env` before running any `make` command. For unit tests the placeholder values are fine; acceptance tests need real credentials.
- **`go generate ./...` (doc generation) may fail** with a template error involving `lightdash_authenticated_user`. This is a pre-existing issue; it does not affect the core build or tests. When you need just the build, run `go build -v ./` directly.
- **`gosec` is installed to `~/go/bin/`.** Ensure `~/go/bin` is on `PATH`.
- **`pre-commit` is installed to `~/.local/bin/`.** Ensure `~/.local/bin` is on `PATH`.

### Testing the provider locally

1. Build + install: `go install .`
2. Add a `dev_overrides` block to `~/.terraformrc` pointing `ubie-oss/lightdash` to your `$GOPATH/bin`.
3. Run `terraform plan` in a directory with a Lightdash provider config. The provider will load from the local binary. A real Lightdash API token is needed for plans/applies to succeed.
