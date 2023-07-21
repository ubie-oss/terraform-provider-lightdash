# Set up environment to develop the provider locally

## Add the provider to your terraform configuration
We have to acc the configuration to install the local provider at `~/.terraformrc`.

```
provider_installation {

  dev_overrides {
      "github.com/ubie-oss/terraform-provider-lightdash" = "local/path/to/provider"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

## Publish the provider to the Terraform Registry

We need some steps to publish the provider to the Terraform Registry using GitHub Actions.
The original creator manually set them up.
For instance, we need to:

- Register a PGP key to the Terraform Registry
- Register GtiHub Actions' secrets

The ofitial tutorials describes the steps in detail.

- [Release and Publish a Provider to the Terraform Registry](https://developer.hashicorp.com/terraform/tutorials/providers/provider-release-publish)
