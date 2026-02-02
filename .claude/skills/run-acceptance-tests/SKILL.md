---
name: run-acceptance-tests
description: Guide the execution of Lightdash Terraform provider acceptance tests, ensuring correct environment setup and targeted test execution.
---

# Run Acceptance Tests

## Purpose

Standardize the execution of acceptance tests (integration tests against a live Lightdash API) in this project. This skill ensures that the environment is correctly configured before running potentially slow or impactful tests.

## Prerequisites & Mandatory Validation

Acceptance tests (setting `TF_ACC=1`) require a live Lightdash instance and specific environment variables.

### 1. Verify Environment (.env file)

Before running any acceptance tests, you **MUST** verify the existence and content of the `.env` file in the project root.

- **Check Existence**: Verify if `.env` exists.
- **Check Variables**: Ensure the following variables are set with valid values:
  - `LIGHTDASH_URL`: The base URL of the Lightdash instance (e.g., `https://app.lightdash.cloud`).
  - `LIGHTDASH_API_KEY`: A valid API key for authentication.
  - `LIGHTDASH_PROJECT`: The UUID of the project to use for testing.

**If `.env` is missing or incomplete:**

- Guide the user to copy `.env.template` to `.env`: `cp .env.template .env`
- Instruct the user to fill in the required values before proceeding.

## Workflow

### 1. Running Acceptance Tests (Standard)

Use this for general verification of the provider's resources and data sources against a live API.

- **Command**: `make testacc`
- **Details**: This runs all tests in `internal/provider/...` with `TF_ACC=1`.
- **Warning**: This can be slow and may create/modify resources in the specified Lightdash project.

### 2. Targeted Acceptance Testing (Recommended)

To save time and focus on specific changes, run only relevant test cases.

- **Command**: `make testacc TESTARGS="-run <Pattern>"`
- **Example (Resource)**: `make testacc TESTARGS="-run TestAccResourceSpace"`
- **Example (Data Source)**: `make testacc TESTARGS="-run TestAccDataSourceProject"`
- **Example (Specific Test Case)**: `make testacc TESTARGS="-run TestAccResourceSpace/create_space"`

## Distinction from Unit Tests

- **Unit Tests**: Run via `make test`. They are fast, do not require a live API, and MUST NOT set `TF_ACC=1`.
- **Acceptance Tests**: Run via `make testacc`. They are slower, require a live API, and MUST set `TF_ACC=1`.

## Troubleshooting

- **"Context deadline exceeded"**: Increase timeout via `TESTARGS="-timeout 120m"`.
- **Authentication Errors**: Verify `LIGHTDASH_API_KEY` and `LIGHTDASH_URL` in `.env`.
- **Resource Cleanup**: If a test fails, you might need to manually delete resources in the Lightdash UI or via the API if they weren't cleaned up automatically.
