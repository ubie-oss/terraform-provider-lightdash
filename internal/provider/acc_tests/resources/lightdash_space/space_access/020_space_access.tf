#######################################################################
# Groups
#######################################################################
data "lightdash_organization" "test" {
}

resource "lightdash_group" "space_access__test_group_1" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "Acceptance Test Group 1"
  members           = []
}

resource "lightdash_group" "space_access__test_group_2" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "Acceptance Test Group 2"
  members           = []
}

resource "lightdash_group" "space_access__test_group_3" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "Acceptance Test Group 3"
  members           = []
}

#######################################################################
# Spaces
#######################################################################
resource "lightdash_space" "space_access__test_space_1" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Space 1 (Acceptance Test: space_access)"
  is_private          = true
  deletion_protection = false

  // Add group access
  group_access {
    group_uuid = lightdash_group.space_access__test_group_1.group_uuid
    space_role = "admin"
  }

  group_access {
    group_uuid = lightdash_group.space_access__test_group_2.group_uuid
    space_role = "editor"
  }

  group_access {
    group_uuid = lightdash_group.space_access__test_group_2.group_uuid
    space_role = "viewer"
  }
}

resource "lightdash_space" "space_access__test_space_2" {
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Space 2 (Acceptance Test: space_access)"
  is_private          = false
  deletion_protection = false

  // Remove group access
}
