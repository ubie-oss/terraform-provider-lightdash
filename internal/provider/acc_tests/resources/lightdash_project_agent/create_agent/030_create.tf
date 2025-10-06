data "lightdash_organization" "test" {
}

resource "lightdash_project_agent" "test" {
  organization_uuid       = data.lightdash_organization.test.organization_uuid
  project_uuid            = data.lightdash_project.test.project_uuid
  name                    = "Test Agent Updated"
  instruction             = "You are an updated helpful AI assistant for data analysis and insights."
  enable_data_access      = true
  enable_self_improvement = false
  deletion_protection     = false
}

output "is_agent_uuid_set" {
  value = (
    lightdash_project_agent.test.agent_uuid != null
    && length(lightdash_project_agent.test.agent_uuid) > 0
  )
}

data "lightdash_project_agent" "test" {
  project_uuid = data.lightdash_project.test.project_uuid
  agent_uuid   = lightdash_project_agent.test.agent_uuid
}
