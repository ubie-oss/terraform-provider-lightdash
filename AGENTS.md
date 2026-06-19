# AGENTS.md

**Primary agent configuration** for this repo. Read this first for shared norms; do not copy long sections into other instruction files unless a tool cannot read `AGENTS.md`. Which tools load it and how: see the table below.

## How agents discover instructions

| Tool                            | Notes                                                                                                           | Official reference                                                                                                                                                                                                                                  |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Cursor**                      | Root and nested `AGENTS.md`; may combine with [`.cursor/rules/`](.cursor/rules/) and other agent config.        | [Rules / AGENTS.md](https://cursor.com/docs/rules)                                                                                                                                                                                                  |
| **OpenAI Codex**                | Instruction chain from `~/.codex` then repo root → cwd; merges `AGENTS.md` / `AGENTS.override.md` (later wins). | [Custom instructions with AGENTS.md](https://developers.openai.com/codex/guides/agents-md/)                                                                                                                                                         |
| **Claude Code**                 | [CLAUDE.md](CLAUDE.md) should start with `@AGENTS.md`; Claude-only notes stay in `CLAUDE.md`.                   | [Import additional files](https://code.claude.com/docs/en/memory#import-additional-files), [CLAUDE.md](https://docs.anthropic.com/en/docs/claude-code/claude-md), [.claude directory](https://code.claude.com/docs/en/claude-directory)             |
| **Gemini CLI**                  | [`.gemini/settings.json`](.gemini/settings.json) sets `context.fileName` to include `AGENTS.md`.                | [Context files](https://geminicli.com/docs/cli/gemini-md/)                                                                                                                                                                                          |
| **GitHub Copilot coding agent** | `AGENTS.md` in repo (nested files: nearest wins).                                                               | [Custom instructions](https://docs.github.com/en/copilot/customizing-copilot/adding-custom-instructions-for-github-copilot), [changelog](https://github.blog/changelog/2025-08-28-copilot-coding-agent-now-supports-agents-md-custom-instructions/) |

### Codex and optional extras

- **Size:** Codex concatenates discovered docs (~32 KiB combined by default). If you outgrow that, add **nested `AGENTS.md`** in subtrees (Codex and Cursor).
- **Optional files:** `.github/copilot-instructions.md` or `.github/instructions/*.instructions.md` only if a Copilot surface does not use `AGENTS.md`. **`.codex/agents/*.toml`** for Codex worker configs ([multi-agent](https://developers.openai.com/codex/multi-agent/))—not a substitute for this file.

### Repo paths (not obvious from the table)

- **`.claude/`** — Claude Code skills, agents, rules, settings. `.claude/agents/*.md` is also a Cursor subagent path; same name in `.cursor/agents/` wins.
- **Skills:** Cursor auto-discovers [`.cursor/skills/`](https://cursor.com/docs/skills) or `.agents/skills/` only. Reference [`.claude/skills/`](.claude/skills/) with `@.claude/skills/...` (or mirror into `.cursor/skills/` if you want auto-discovery).

## Overview

This is a Terraform provider for Lightdash, written in Go. It is a single-product repo (not a monorepo). There are no local services to run — the provider binary communicates with a remote Lightdash API instance.

## Key commands

All standard dev commands are defined in `GNUmakefile`:

- **Unit tests:** `make test` (runs `go test` inside `internal/`, skipping acceptance tests)
- **Build:** `go build -v ./` (the plain Go build; `make build` also runs `gen-docs`, `go-tidy`, `gosec`, `deadcode`)
- **Lint:** `trunk check --all` (the `make lint` target also runs `pre-commit run --all-files`)
- **Format:** `make format` (runs `go fmt` + `trunk fmt --all`)
- **Acceptance tests:** `make testacc` (requires a live Lightdash instance; see `CONTRIBUTING.md`)

## Gotchas

- **`.env` file is required.** The `GNUmakefile` has `include .env` at line 19, which is unconditional. If `.env` does not exist, every `make` target fails — even `make test`. Copy `.env.template` to `.env` before running any `make` command. For unit tests the placeholder values are fine; acceptance tests need real credentials.
- **`go generate ./...` (doc generation) may fail** with a template error involving `lightdash_authenticated_user`. This is a pre-existing issue; it does not affect the core build or tests. When you need just the build, run `go build -v ./` directly.
- **`gosec` is installed to `~/go/bin/`.** Ensure `~/go/bin` is on `PATH`.
- **`pre-commit` is installed to `~/.local/bin/`.** Ensure `~/.local/bin` is on `PATH`.

## Testing the provider locally

1. Build + install: `go install .`
2. Add a `dev_overrides` block to `~/.terraformrc` pointing `ubie-oss/lightdash` to your `$GOPATH/bin`.
3. Run `terraform plan` in a directory with a Lightdash provider config. The provider will load from the local binary. A real Lightdash API token is needed for plans/applies to succeed.

More setup and acceptance-test detail: [CONTRIBUTING.md](CONTRIBUTING.md).

## Learned User Preferences

- When implementing an attached plan, do not edit the plan file; use existing todos instead of recreating them.
- For high-impact credentials (e.g. OAuth clients), prefer `deletion_protection` matching `lightdash_space`: Terraform-only delete guard with import default `true`.
- Do not add acceptance tests for `lightdash_project_role_member` or `lightdash_organization_role_member`; they mutate real user org/project access and are unsafe on production-like instances.
- Avoid unit or integration tests that call live APIs in ways that mutate existing org/project access; prefer offline JSON fixtures and pure service logic tests.
- On `lightdash_project_role_member` Update, pass `send_email=false` to the v2 assign call; v1 update never sent notification emails, so only Create should honor Terraform `send_email`.

## Learned Workspace Facts

- In git worktrees, `tfplugindocs` may infer the provider name from the directory basename (e.g. `gv41`); keep `--provider-name lightdash` on the `main.go` `go:generate` line or `make gen-docs` fails template rendering.
- Embedded markdown under `internal/provider/docs/` should be plain descriptive prose only—do not paste generated schema bullet lists; schema-like embedded content corrupts registry docs (nested frontmatter, trailing-whitespace `make lint` failures).
- Shared Terraform string Set/List conversion helpers belong in `internal/provider/utils.go`, not in individual resource or data-source files.
- v2 `RoleAssignment` is the shared schema for org and project `.../roles/assignments` list endpoints; it is IAM-assignment-centric and not a drop-in replacement for v1 member-directory APIs (`GET /api/v1/org/users`, `GET /api/v1/projects/{id}/access`).
- Before implementing v2 API models/clients, verify GET response shapes with live calls against `.env`; swagger alone is insufficient.
- Wire v2 role IAM through dedicated `RoleService`; do not replace v1 HTTP clients inside member-directory services (`OrganizationMembersService`, `ProjectService`).
- `lightdash_project_role_member` Create, Read, Update, and Delete use v2 assignments via `RoleService`; v1 is used only for org member email lookup on Create/Read.
- `GetRoleService` in the provider must be a singleton (same pattern as `OrganizationMembersService`) so org role-catalog caching is shared across resources.
- `lightdash_project_group_accesses` reads via v2 project role assignments (group filter); org and project member list data sources stay on v1 directory APIs.
- Live v2 payloads use slug role IDs (e.g. `"editor"`) for system roles, not UUIDs—role resolution must match catalog `roleUuid` slugs and normalized display names.
