---
name: verify-and-fix
description: Run formatting, linters, unit tests, and build checks, then fix violations.
---

# Verify and Fix

## Purpose

Automate the process of ensuring code quality, functional correctness, and buildability within the Terraform Provider Lightdash project. This skill provides a comprehensive local verification loop using `trunk`, `make`, and `pre-commit`.

## Safety & Restrictions

- **NO Acceptance Tests**: Do NOT run `make testacc` or any command that sets `TF_ACC=1`. This skill is intended for fast, local feedback loops (linting, building, unit testing) and should not interact with live infrastructure.

## Workflow

### 1. Formatting

Start by ensuring consistent code formatting across the repository.

- **Command**: `trunk fmt --all`
- **Action**: Run this command to automatically format all files. If any changes are made, they will be reflected in the working directory.

### 2. Linting

Execute the project's linting suite to identify violations.

- **Command**: `make lint`
- **Details**: This command runs `trunk check --all` and `pre-commit run --all-files`.
- **Analysis**: Review the output for any errors or warnings.

### 3. Unit Testing

Execute the project's unit tests to ensure functional correctness.

- **Command**: `make test`
- **Details**: This runs `go test` on the internal packages without setting `TF_ACC`.
- **Action**: Verify that all tests pass.

### 4. Automated Fixing

If linter violations are detected, attempt to resolve them automatically.

- **Condition**: Only if `make lint` reports fixable issues.
- **Command**: `trunk check --all --fix`
- **Action**: This will apply automated fixes provided by the linters integrated into Trunk.

### 5. Build Verification

Ensure the project compiles successfully and passes security/deadcode checks.

- **Command**: `make build`
- **Details**: This triggers `gen-docs`, `go-tidy`, `gosec`, `deadcode`, and finally `go build`.
- **Action**: Verify that the build completes without errors.

### 6. Iterative Manual Fixes

If issues remain after automated attempts, proceed with manual intervention.

- **Analyze**: Examine the specific error messages from `trunk check`, `make test`, or `make build`.
- **Fix**: Use the `Edit` tool to address the root causes of the violations in the source code.
- **Verify**: Re-run the loop starting from **Step 2 (Linting)** until all checks pass and the build is successful.

## Termination Criteria

- `make lint` returns no errors.
- `make test` returns no failures.
- `make build` completes successfully.
- No remaining linter violations are reported by `trunk check --all`.

## Examples

### Scenario: Fixing a Bug Found by Unit Tests

1. **Format**: `trunk fmt --all` (Success).
2. **Lint**: `make lint` (Success).
3. **Unit Testing**: `make test` (Fails in `internal/lightdash/api/utils_test.go`).
4. **Manual Fix**:
   - Analyze the test failure.
   - Edit `internal/lightdash/api/utils.go` to fix the logic.
5. **Verify**: Re-run `make test`. (Success).
6. **Build**: `make build`. (Success).
