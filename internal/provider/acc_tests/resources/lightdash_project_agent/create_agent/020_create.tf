data "lightdash_organization" "test" {
}

resource "lightdash_project_agent" "test" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  project_uuid      = data.lightdash_project.test.project_uuid

  version                 = 2
  name                    = "Test Agent Updated"
  instruction             = "You are an updated helpful AI assistant for data analysis and insights."
  tags                    = ["terraform", "updated"]
  enable_data_access      = true
  enable_self_improvement = true
  group_access            = []
  user_access             = []
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
