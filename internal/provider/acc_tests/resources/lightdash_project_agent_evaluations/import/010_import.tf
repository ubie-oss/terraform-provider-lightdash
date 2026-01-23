data "lightdash_organization" "test" {
}

# Create an agent to use for evaluation
resource "lightdash_project_agent" "test_agent" {
  organization_uuid   = data.lightdash_organization.test.organization_uuid
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Test Agent for Evaluation Import"
  description         = "Test Description for Evaluation Import"
  instruction         = "You are a helpful AI assistant for data analysis and evaluation imports."
  deletion_protection = false
  tags                = ["import", "test"]
  enable_data_access  = true
  enable_reasoning    = false
  space_access        = []
}

# Create an evaluation to import
resource "lightdash_project_agent_evaluations" "test_evaluation" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  project_uuid      = data.lightdash_project.test.project_uuid
  agent_uuid        = lightdash_project_agent.test_agent.agent_uuid
  title             = "Test Evaluation for Import"
  description       = "This evaluation is created for import testing"
  prompts = [
    {
      prompt            = "Show me the top 5 customers by sales."
      expected_response = ""
    },
    {
      prompt            = "What are the most popular products?"
      expected_response = ""
    }
  ]
  deletion_protection = false
}
