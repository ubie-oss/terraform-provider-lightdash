---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "lightdash_group Data Source - terraform-provider-lightdash"
subcategory: ""
description: |-
  Lightdash group data source
---

# lightdash_group (Data Source)

Lightdash group data source

## Example Usage

```terraform
data "lightdash_group" "test" {
  organization_uuid = "xxxxx-xxxxxx-xxxx"
  project_uuid      = "xxxxx-xxxxxx-xxxx"
  group_uuid        = "xxxxx-xxxxxx-xxxx"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `group_uuid` (String) UUID of the Lightdash group.
- `organization_uuid` (String) Organization UUID of the Lightdash group.
- `project_uuid` (String) UUID of the Lightdash project.

### Read-Only

- `created_at` (String) Creation timestamp of the Lightdash group.
- `id` (String) Data source identifier
- `name` (String) Name of the Lightdash group.