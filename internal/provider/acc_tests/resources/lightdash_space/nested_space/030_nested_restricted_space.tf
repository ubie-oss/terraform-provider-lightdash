data "lightdash_organization" "test" {}

resource "lightdash_space" "nested_restricted_root" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Nested Restricted Root (Acceptance Test)"
  is_private          = true
  deletion_protection = false
}

resource "lightdash_group" "nested_restricted_group" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "Nested Restricted Group"
  members           = []
}

resource "lightdash_space" "nested_restricted_child" {
  project_uuid        = data.lightdash_project.test.project_uuid
  parent_space_uuid   = lightdash_space.nested_restricted_root.space_uuid
  name                = "Nested Restricted Child (Acceptance Test)"
  is_private          = true
  deletion_protection = false

  group_access {
    group_uuid = lightdash_group.nested_restricted_group.group_uuid
    space_role = "editor"
  }
}
