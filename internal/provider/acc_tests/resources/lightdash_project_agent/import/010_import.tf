data "lightdash_organization" "test" {
}

# Create an agent to import
resource "lightdash_project_agent" "test_agent" {
  organization_uuid       = data.lightdash_organization.test.organization_uuid
  project_uuid            = data.lightdash_project.test.project_uuid
  name                    = "Test Agent for Import"
  instruction             = "You are a helpful AI assistant for data analysis and imports."
  tags                    = ["import", "test"]
  enable_data_access      = true
  enable_self_improvement = true
  deletion_protection     = false
}
