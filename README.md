# Terraform Provider Lightdash

[![Release](https://github.com/ubie-oss/terraform-provider-lightdash/actions/workflows/release.yml/badge.svg)](https://github.com/ubie-oss/terraform-provider-lightdash/actions/workflows/release.yml)
[![Tests](https://github.com/ubie-oss/terraform-provider-lightdash/actions/workflows/test.yml/badge.svg)](https://github.com/ubie-oss/terraform-provider-lightdash/actions/workflows/test.yml)

The Terraform provider for Lightdash enables you to manage your Lightdash resources with ease.

## Usage Guide

Below is a step-by-step example demonstrating how to assign the editor role to a user at the project level using Terraform.

```hcl
# Set up the Lightdash provider
terraform {
  required_providers {
    lightdash = {
      source  = "ubie-oss/lightdash"
      version = "0.4.2"
    }
  }
}

provider "lightdash" {
  host  = "https://app.lightdash.cloud"  # Replace with your Lightdash host
  token = var.personal_access_token       # Replace with your personal access token
}

# Retrieve details about the organization
data "lightdash_organization" "my_organization" {}

# Retrieve details about a specific user within the organization
data "lightdash_organization_member" "test_user" {
  organization_uuid = data.lightdash_organization.my_organization.organization_uuid
  email             = "test-user@example.com"
}

# Retrieve details about a specific project
data "lightdash_project" "jaffle_shop" {
  project_uuid = "xxxx-xxxx-xxxx"  # Replace with your project's UUID
}

# Assign the editor role to the user for the specified project
resource "lightdash_project_role_member" "test" {
  project_uuid = data.lightdash_project.jaffle_shop.project_uuid
  user_uuid    = data.lightdash_organization_member.test_user.user_uuid
  role         = "editor"
}
```

## Developer Guide

### Prerequisites

- Terraform version 1.1 or higher
- Go version 1.19 or higher

### Building the Provider

To build the Lightdash provider from source:

1. Clone the repository to your local machine.
2. Navigate to the repository directory.
3. Execute the following command to build the provider:

```shell
go install
```

### Managing Dependencies

The Lightdash provider is built using Go modules. For the latest guidelines on using Go modules, refer to the [Go modules documentation](https://github.com/golang/go/wiki/Modules).

To add a new Go module dependency, for example `github.com/author/dependency`, use the following commands:

```shell
go get github.com/author/dependency
go mod tidy
```

Afterwards, commit the updated `go.mod` and `go.sum` files to your version control system.

### Provider Usage

To use the provider, follow the instructions in the Usage Guide section above.

### Provider Development

If you're interested in contributing to the development of the Lightdash provider, ensure you have Go installed on your system (refer to [Prerequisites](#prerequisites)).

To compile the provider, run `go install`. This will build the provider binary and place it in the `$GOPATH/bin` directory.

To update or generate new documentation, use the `make gen-docs` command.

For running the full suite of Acceptance tests, which create actual resources and may incur costs, execute:

```shell
make testacc
```

## How to Contribute

Contributions to the `terraform-provider-lightdash` are welcome! Check out our [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to get involved.

## Cautions

### Group-level space access control is not recommended

We advise against using group-level space access control at this time, as it requires further improvements to ensure optimal functionality.

<https://github.com/lightdash/lightdash/issues/10883>
