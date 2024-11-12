# Contributing to `terraform-provider-lightdash`

The `terraform-provider-lightdash` project is open source and we warmly welcome contributions from the community.
Whether you're fixing bugs, adding new features, or improving the documentation, your efforts will help enhance the Lightdash provider for everyone.

The provider is built using the Terraform Plugin SDK.
For those interested in contributing, we recommend reviewing the [Custom Framework Providers tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework) provided by HashiCorp.
This resource offers an excellent overview of the SDK and serves as a solid foundation for understanding how to enhance this provider.

## Setting Up Your Local Development Environment

### Prerequisites

Before you begin, ensure you have the following prerequisites installed:

- Terraform: Use `tfenv install` to install the specific Terraform version required, as indicated in the `.terraform-version` file.
- GNU Make: Utilize the `make` command to facilitate building and testing the provider. Targets are defined in the [GNUmakefile](./GNUmakefile).
- [Trunk](https://docs.trunk.io/) (Optional): Trunk is a developer experience (DevEx) toolkit that enables you to ship code quickly while maintaining the necessary guardrails for excellent eng teams. If you use macOS, `brew install trunk-io` enables you to install Trunk.

The subsequent commands enable us to set up the local development environment.

```shell
make setup-dev
```

### Configuring the Provider for Local Development

To use the local version of the `terraform-provider-lightdash`, you must update your Terraform configuration. Add the local provider to your `~/.terraformrc` file to override the default provider source location.

```terraform
provider_installation {

  dev_overrides {
      "github.com/ubie-oss/terraform-provider-lightdash" = "local/path/to/provider"
  }

# For all other providers, install them directly from their origin provider

# registries as normal. If you omit this, Terraform will _only_ use

# the dev_overrides block, and so no other providers will be available

  direct {}
}
```

### Building and Testing the Provider Locally

To ensure the quality and functionality of your changes, it's essential to build and test the provider before submitting a contribution.
Follow the steps below to compile the provider and run the automated test suite.

```shell
# Compile the provider binary from source
make build

# Execute the automated test suite to verify your changes
make test
```

### Integration Testing

To validate the integration of your changes with a live Lightdash instance, follow these steps:

1. Obtain an API token from your Lightdash instance to authenticate requests.
2. Optionally, set up a Lightdash project specifically for testing the provider's functionality.
3. Generate a `.tfvars` file using the provided template at [integration_tests/testing.tfvars.template](./integration_tests/testing.tfvars.template).
4. Rebuild the provider binary to include your latest changes by running `make build`.
5. Execute the integration tests to ensure your changes interact correctly with Lightdash.
6. After testing, clean up by destroying the test resources to avoid lingering infrastructure.

Execute the following commands within the `integration_tests` directory to perform the integration testing:

```shell
cd integration_tests

# Review the planned changes for the test resources
terraform plan -var-file="testing.tfvars"

# Apply the changes to create the test resources
terraform apply -var-file="testing.tfvars"

# Remove the test resources after testing is complete
terraform destroy -var-file="testing.tfvars"
```

## Publishing the Provider to the Terraform Registry

### Account Setup for Terraform Registry

Before you can publish your provider, you must set up an account on the Terraform Registry. This is a prerequisite for the subsequent steps in the publishing process.

To publish the provider to the Terraform Registry, a series of steps must be followed using GitHub Actions. These steps were initially configured manually by the original author.
For instance, we need to:

- Register a PGP key to the Terraform Registry
- Register GitHub Actions' secrets

The official tutorials describes the steps in detail.

- [Release and Publish a Provider to the Terraform Registry](https://developer.hashicorp.com/terraform/tutorials/providers/provider-release-publish)

### Publish a New Release

All we have to do is to create a new release on GitHub.
The GitHub Actions will automatically publish the provider to the Terraform Registry.
