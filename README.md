# Terraform Provider Lightdash

A terraform provider for Lightdash.

## How to use

Here is an example to grant the editor role at project level to a user.

```
# Configure the Lightdash provider
terraform {
  required_providers {
    lightdash = {
      source = "registory.terraform.io/ubie-oss/lightdash"
      version = "0.0.1"
    }
  }
}

provider "lightdash" {
  host  = "https://app.lightdash.cloud"  # Use your host
  token = var.personal_access_token
}

# Get the organization data source
data "lightdash_organization" "my_organization" {}

# Get the user data source
data "lightdash_organization_member" "test_user" {
  organization_uuid = data.lightdash_organization.my_organization.organization_uuid
  email = "test-user@example.com"
}

# Get the project data source
data "lightdash_project" "jaffle_shop" {
  project_uuid = "xxxx-xxxx-xxxx"
}

# Grant the editor role of the project to the user
resource "lightdash_project_role_member" "test" {
  project_uuid = data.lightdash_project.jaffle_shop.project_uuid
  user_uuid    = data.lightdash_organization_member.test_user.user_uuid
  role         = "editor"
}
```

## Development

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.1
- [Go](https://golang.org/doc/install) >= 1.19

### Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

### Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

### Using the provider

Fill this in for each provider

### Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
