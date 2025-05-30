---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "lightdash_organization_members_by_emails Data Source - terraform-provider-lightdash"
subcategory: ""
description: |-
  Fetches Lightdash organization members filtered by a list of emails. This data source retrieves a list of members within a Lightdash organization whose email addresses are included in the provided list. It is useful for obtaining details (user UUID, email, and organization role) for a specific subset of organization members. The results are returned as a list of members, sorted by user UUID.
---

# lightdash_organization_members_by_emails (Data Source)

Fetches Lightdash organization members filtered by a list of emails. This data source retrieves a list of members within a Lightdash organization whose email addresses are included in the provided list. It is useful for obtaining details (user UUID, email, and organization role) for a specific subset of organization members. The results are returned as a list of members, sorted by user UUID.

## Example Usage

```terraform
data "lightdash_organization_members_by_emails" "test" {
  emails = [
    "test@test.com",
    "test2@test.com",
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `emails` (List of String, Sensitive) A list of email addresses to filter the organization members by. Only members with an email in this list will be returned.

### Read-Only

- `id` (String) Identifier of the data source, computed as `organizations/<organization_uuid>/users`.
- `members` (Attributes List) A list of organization members matching the provided emails, sorted by user UUID. (see [below for nested schema](#nestedatt--members))
- `organization_uuid` (String) The UUID of the organization the members belong to.

<a id="nestedatt--members"></a>
### Nested Schema for `members`

Read-Only:

- `email` (String) The email address of the Lightdash user.
- `role` (String) The organization role of the Lightdash user (e.g., `viewer`, `editor`, `admin`).
- `user_uuid` (String) The UUID of the Lightdash user.
