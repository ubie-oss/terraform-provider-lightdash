data "lightdash_organization" "test" {}

resource "lightdash_group" "nested_restricted_group" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "Nested Restricted Group (Acceptance Test)"
  members           = []
}

resource "lightdash_space" "parent1" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Parent 1 (Acceptance Test)"
  is_private          = true
  deletion_protection = false
}

resource "lightdash_space" "parent2" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Parent 2 (Acceptance Test)"
  is_private          = true
  deletion_protection = false
}

resource "lightdash_space" "test_space" {
  project_uuid        = data.lightdash_project.test.project_uuid
  parent_space_uuid   = lightdash_space.parent1.space_uuid
  name                = "Test Space (Acceptance Test)"
  is_private          = true
  deletion_protection = false

  group_access {
    group_uuid = lightdash_group.nested_restricted_group.group_uuid
    space_role = "editor"
  }
}
