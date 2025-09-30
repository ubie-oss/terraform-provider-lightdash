resource "lightdash_project_agent_evaluations" "test_evaluations" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  project_uuid      = data.lightdash_project.test.project_uuid
  agent_uuid        = lightdash_project_agent.test_agent.agent_uuid
  title             = "Test Evaluations"
  description       = "Test Evaluations Description"
  prompts = [
    {
      prompt = "What are the sales trends over time?"
    },
  ]
}
