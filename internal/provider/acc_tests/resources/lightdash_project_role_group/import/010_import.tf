data "lightdash_organization" "test" {
}

resource "lightdash_group" "test_group" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "test-grant-project-role-group (Acceptance Test - import)"
  members           = []
}

resource "lightdash_project_role_group" "test_project_role_group" {
  project_uuid = data.lightdash_project.test.project_uuid
  group_uuid   = lightdash_group.test_group.group_uuid
  role         = "viewer"
}
