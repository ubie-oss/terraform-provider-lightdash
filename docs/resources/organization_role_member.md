---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "lightdash_organization_role_member Resource - terraform-provider-lightdash"
subcategory: ""
description: |-
  Lightash the role of a member at organization level
---

# lightdash_organization_role_member (Resource)

Lightash the role of a member at organization level

## Example Usage

```terraform
resource "lightdash_organization_role_member" "test" {
  user_uuid = "xxxx-xxx-xxx"
  role      = "editor"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `organization_uuid` (String) Lightdash organization UUID
- `role` (String) Lightdash organization role
- `user_uuid` (String) Lightdash user UUID

### Read-Only

- `email` (String) Lightdash user UUID
- `id` (String) Resource identifier
- `last_updated` (String) Timestamp of the last Terraform update of the space.

## Import

Import is supported using the following syntax:

```shell
# Organization role members can be imported by specifying the resource identifier.
terraform import lightdash_organization_role_member.example "organizations/${organization_uuid}/users/${space-uuid}"
```
