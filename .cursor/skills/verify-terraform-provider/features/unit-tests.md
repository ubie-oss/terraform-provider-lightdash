# Unit tests (offline)

Practitioners and CI rely on `make test` to guard provider logic without touching a live Lightdash org. This feature runs that path and captures the log as proof.

## Sub-features

- **unit-tests-default** — `make test` with `TF_ACC` unset (acceptance tests skipped inside packages).
- **env-file-present** — GNUmakefile `include .env`; placeholders from `.env.template` are sufficient.

## How to get to it (user POV)

- Clone the repo, copy `.env.template` to `.env`, run `make test` from the root (see `AGENTS.md`, `CONTRIBUTING.md`).

## Driving it with verify-terraform-provider

Preconditions: `scripts/launch.sh` optional for this feature (doctor checks `.env` only); evidence dir created via `common.sh`.

- **Run unit tests** — `.cursor/skills/verify-terraform-provider/scripts/drive.sh unit-tests` — Observable: exit code `0`, `unit-tests.log` under `/opt/cursor/artifacts/verify-terraform-provider/${VERIFY_RUN_ID}/` ending with passing `go test` output.

## Gotchas

- **`make test` fails without `.env`** even though unit tests do not need real credentials.
- Do not set `TF_ACC=1`; that enables acceptance tests in some packages.
- Full `make build` runs docgen and may fail on known `lightdash_authenticated_user` template issues; unit tests use `make test` only.
