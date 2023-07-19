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
