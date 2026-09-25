---
name: verify-terraform-provider
description: "Drive the Lightdash Terraform provider the way a practitioner does: build/install the plugin, point Terraform CLI at it via dev_overrides, run unit tests and offline terraform validate/schema against examples. Use after provider or schema changes, before claiming a feature works end-to-end, or when you need proof without running acceptance tests or mutating live org/project access."
---

# Verify Terraform Provider (Lightdash)

This skill exercises **Terraform CLI + a locally installed provider binary** (`go install`, `dev_overrides` in a disposable `terraformrc`). It does **not** start a local Lightdash server. Primary proof is **unit tests** and **offline `terraform validate` / `providers schema`**. Live **`terraform plan`** against read-only data sources requires a valid Lightdash host and API token (see gaps in feature map).

## Launch

Build and install the provider once per verification run, and write a **disposable** CLI config that overrides the registry provider with your local binary.

From the repository root:

```bash
export VERIFY_RUN_ID="verify-$(date -u +%Y%m%dT%H%M%SZ)-$$"
export PATH="$HOME/.local/bin:$PATH:$(go env GOPATH)/bin"
.cursor/skills/verify-terraform-provider/scripts/launch.sh
```

**Ready when:** the script prints `Launch OK` and `${GOPATH}/bin/terraform-provider-lightdash` exists.

**Teardown:** `scripts/cleanup.sh` removes scratch dirs under `/tmp/lightdash-tf-verify-$VERIFY_RUN_ID` and run-state under `.cursor/skills/verify-terraform-provider/.run-state/$VERIFY_RUN_ID`. It does **not** remove evidence.

**Prerequisites the launch script ensures:**

- Copy `.env.template` → `.env` if missing (required for `make test`; placeholder values are fine for unit tests).
- Terraform **1.14.0** on `PATH` (see repo `.terraform-version` and `CONTRIBUTING.md`). Cloud agents may install the binary to `~/.local/bin`.

**Isolation:** Each run uses its own `VERIFY_RUN_ID`, scratch directory, and `TF_CLI_CONFIG_FILE` pointing at a generated `terraformrc`. Do not reuse another agent's `VERIFY_RUN_ID`. Never run `make testacc`, `TF_ACC=1`, or `integration_tests/` apply against production-like orgs without explicit approval.

## Doctor

Read-only check that the toolchain and local override are worth driving:

```bash
.cursor/skills/verify-terraform-provider/scripts/doctor.sh
```

**Pass criteria:** exit code `0` and `Doctor: READY`. Confirms `.env` exists, `go` and `terraform` run, `terraform-provider-lightdash` is installed, and `dev_overrides` for `ubie-oss/lightdash` is present in the run's `terraformrc`.

Run doctor whenever launch was skipped, after a failed drive, or if Terraform loads the wrong provider build.

## Drive

Read the feature index first: [`features/README.md`](features/README.md). Drive **one feature per proof** unless you are explicitly regression-testing several entry points.

```bash
export TF_CLI_CONFIG_FILE="${VERIFY_STATE_DIR}/terraformrc"   # set by launch; doctor uses the same run_id
.cursor/skills/verify-terraform-provider/scripts/drive.sh <feature-id>
```

**Feature IDs:** `unit-tests`, `provider-local-install`, `terraform-validate-organization-ds`, `provider-schema-json`, `resource-space-schema`.

**Harness conventions:**

- Provider source address in HCL: `ubie-oss/lightdash` (registry: `registry.terraform.io/ubie-oss/lightdash`).
- Offline validate workspaces use a non-routable host and a dummy token; **`terraform validate` and `providers schema` do not call the Lightdash API**. **`terraform plan` / `apply` always configure the provider and validate the token** against the API — use only with real credentials in a disposable workspace, never for role-member mutations.
- Prefer example shapes under `examples/data-sources/` and `examples/resources/` when composing scratch configs.
- Do **not** run `make testacc` or acceptance tests that change org/project IAM.

## Evidence

Capture **command transcripts and structured output** under:

```text
/opt/cursor/artifacts/verify-terraform-provider/${VERIFY_RUN_ID}/
```

Each drive appends to `run.meta` and writes feature-specific logs (e.g. `unit-tests.log`, `terraform-validate-organization-ds.log`, `provider-schema.json`).

**Proof standards:**

- Exercise the **user path**: Terraform CLI invoking the **local** provider binary via dev_overrides, not only `go test` mocks in isolation (unit-tests feature is still valid as the project's standard fast gate).
- Capture **both** the command and **observable output** (exit code, validate success line, schema JSON size/hash).
- For schema features, retain a JSON artifact, not only "validate succeeded".
- **`terraform plan` with read-only data sources** proves API read paths only when a **valid** `host` + `token` are supplied; record the plan excerpt showing data source refresh. Without credentials, report the feature as **blocked** with the doctor/validate path used instead — do not claim plan verified.

Cleanup must **never** delete `/opt/cursor/artifacts/verify-terraform-provider/`.

## Cleanup

After capturing evidence (including failed attempts):

```bash
.cursor/skills/verify-terraform-provider/scripts/cleanup.sh
```

Removes:

- `/tmp/lightdash-tf-verify-${VERIFY_RUN_ID}/` (Terraform scratch)
- `.cursor/skills/verify-terraform-provider/.run-state/${VERIFY_RUN_ID}/` (generated `terraformrc`, launch meta)

Preserves:

- `/opt/cursor/artifacts/verify-terraform-provider/${VERIFY_RUN_ID}/`

**Never** kill processes by name. This provider has no long-lived server; if you started a background Terraform command with a known PID, stop that PID only.

Confirm evidence survived:

```bash
test -d "/opt/cursor/artifacts/verify-terraform-provider/${VERIFY_RUN_ID}" && ls -la "/opt/cursor/artifacts/verify-terraform-provider/${VERIFY_RUN_ID}"
```

## Helpers

All scripts live in `.cursor/skills/verify-terraform-provider/scripts/` (executable).

| Script       | Purpose                                                                   |
| ------------ | ------------------------------------------------------------------------- |
| `common.sh`  | `VERIFY_RUN_ID`, evidence/scratch paths, `PATH`, `.env` helper            |
| `launch.sh`  | `go install .`, write dev_overrides `terraformrc`                         |
| `doctor.sh`  | Read-only preflight                                                       |
| `drive.sh`   | Run one feature id, write evidence                                        |
| `cleanup.sh` | Tear down scratch/state; keep evidence                                    |
| `prove.sh`   | `launch → doctor → drive <feature> → cleanup` and assert evidence remains |

**One-shot proof (default feature: organization data source validate):**

```bash
export VERIFY_RUN_ID="verify-$(date -u +%Y%m%dT%H%M%SZ)-$$"
.cursor/skills/verify-terraform-provider/scripts/prove.sh terraform-validate-organization-ds
```

**Related repo skills:** `@.claude/skills/verify-and-fix` for lint/build/test loops; this skill adds Terraform CLI + local provider verification.

**Maintenance:** When routes, resources, or examples change, update `features/` and re-run `prove.sh` for the affected feature id. Consider `/maintain-verification-skill` (pstack) for ongoing map hygiene.
