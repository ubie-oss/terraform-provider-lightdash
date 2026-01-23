# Testing Setup Reference

This document explains the common testing utilities used in the Lightdash Terraform provider acceptance tests.

## `isIntegrationTestMode()`

- Defined in `internal/provider/provider_test.go`.
- Checks if the environment variable `TF_ACC` is set to `1`.
- Used to skip acceptance tests during normal `go test ./...` runs unless explicitly enabled.

## `testAccPreCheck(t)`

- Ensures required environment variables (like `LIGHTDASH_API_KEY` and `LIGHTDASH_HOST`) are set before running acceptance tests.
- Fails the test early if configuration is missing.

## `testAccProtoV6ProviderFactories`

- A map of provider factories used by the `terraform-plugin-testing` framework to instantiate the provider for tests.
- Uses the `New` function from `internal/provider/provider.go`.

## `getProviderConfig()`

- Generates the standard `provider "lightdash" { ... }` block for use in test configurations.
- Reads credentials from environment variables.

## `ReadAccTestResource(pathParts)`

- Utility to read `.tf` files from the `internal/provider/acc_tests` directory.
- Example: `ReadAccTestResource([]string{"resources", "lightdash_space", "010_create.tf"})`.
