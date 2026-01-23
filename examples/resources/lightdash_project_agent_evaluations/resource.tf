resource "lightdash_project_agent_evaluations" "test" {
  organization_uuid = "xxxx-xxxx-xxxx"
  project_uuid      = "xxxx-xxxx-xxxx"
  agent_uuid        = "xxxx-xxxx-xxxx"

  title       = "Test Evaluation"
  description = "Test Description"
  prompts = [
    {
      prompt = "What are the sales trends over time?"
    },
    {
      prompt = "What are the sales trends over time?"
    },
  ]
}
