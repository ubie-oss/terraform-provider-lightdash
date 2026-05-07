data "lightdash_organization" "test" {
}

resource "lightdash_project_agent" "minimal" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  project_uuid      = data.lightdash_project.test.project_uuid

  name        = "Acc Test Minimal Defaults Agent"
  instruction = "You are an acceptance test agent for minimal Terraform configuration."

  deletion_protection = false

  # description and space_access intentionally omitted — provider defaults apply.
}

output "is_minimal_agent_uuid_set" {
  value = (
    lightdash_project_agent.minimal.agent_uuid != null
    && length(lightdash_project_agent.minimal.agent_uuid) > 0
  )
}
