data "lightdash_organization" "test" {
}

resource "lightdash_project_agent" "test" {
  organization_uuid   = data.lightdash_organization.test.organization_uuid
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Test Agent"
  instruction         = "You are a helpful AI assistant for data analysis."
  deletion_protection = true
}

output "is_agent_uuid_set" {
  value = (
    lightdash_project_agent.test.agent_uuid != null
    && length(lightdash_project_agent.test.agent_uuid) > 0
  )
}
