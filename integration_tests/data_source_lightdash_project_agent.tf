data "lightdash_project_agent" "test" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  project_uuid      = data.lightdash_project.test.project_uuid
  agent_uuid        = lightdash_project_agent.test_agent.agent_uuid
}

output "lightdash_project_agent_test" {
  value = data.lightdash_project_agent.test
}
