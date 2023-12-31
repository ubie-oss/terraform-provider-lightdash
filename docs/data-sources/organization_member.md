---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "lightdash_organization_member Data Source - terraform-provider-lightdash"
subcategory: ""
description: |-
  Lightdash organization member data source
---

# lightdash_organization_member (Data Source)

Lightdash organization member data source

## Example Usage

```terraform
data "lightdash_organization_member" "example" {
  email = "test@example.com"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `email` (String) Lightdash user's email.

### Read-Only

- `id` (String) Data source identifier
- `organization_uuid` (String) Lightdash organization UUID.
- `role` (String) Lightdash organization role of the user.
- `user_uuid` (String) Lightdash user UUID.
