resource "lightdash_project_agent" "test_agent" {
  organization_uuid  = data.lightdash_organization.test.organization_uuid
  project_uuid       = data.lightdash_project.test.project_uuid
  name               = "Test Agent"
  instruction        = "test You are a helpful AI assistant for data analysis."
  tags               = ["test", "terraform"]
  enable_data_access = true
  group_access = [
    lightdash_project_role_group.test1.group_uuid,
  ]
  user_access = []
}
