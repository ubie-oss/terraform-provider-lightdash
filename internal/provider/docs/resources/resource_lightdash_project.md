Manages a Lightdash project whose dbt connection type is `none`. This resource does not set warehouse credentials and does not support other dbt connection types.

The Lightdash API has no endpoint that deletes a project. When `deletion_protection` is `false`, `terraform destroy` removes the project from Terraform state and returns a warning that names the project. The project remains in Lightdash and must be deleted in the UI. When `deletion_protection` is `true`, destroy returns an error and the project stays in state. Imported projects default to `deletion_protection = true`.

To drop a project from state without running destroy, replace the resource with a `removed` block and set `lifecycle { destroy = false }`. The `removed` block requires Terraform 1.7 or later.

Changing `name`, `type`, or `dbt_version` replaces the Terraform resource. The previous Lightdash project is left in place.
