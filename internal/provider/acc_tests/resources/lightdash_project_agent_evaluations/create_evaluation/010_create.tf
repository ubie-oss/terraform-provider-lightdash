data "lightdash_organization" "test" {
}

resource "lightdash_project_agent" "test_agent" {
  organization_uuid   = data.lightdash_organization.test.organization_uuid
  project_uuid        = data.lightdash_project.test.project_uuid
  name                = "Test Agent for Evaluation"
  description         = "Test Description for Evaluation"
  instruction         = "You are a helpful AI assistant for data analysis and evaluations."
  deletion_protection = false
  tags                = ["evaluation", "test"]
  enable_data_access  = true
  enable_reasoning    = false
  space_access        = []
}

resource "lightdash_project_agent_evaluations" "test" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  project_uuid      = data.lightdash_project.test.project_uuid
  agent_uuid        = lightdash_project_agent.test_agent.agent_uuid
  title             = "Data Analysis Evaluation"
  description       = "Evaluating the AI assistant's ability to analyze data and provide insights"
  prompts = [
    {
      prompt            = "Show me the top 5 customers by sales."
      expected_response = ""
    },
    {
      prompt            = "What are the most popular products?"
      expected_response = ""
    },
    {
      prompt            = "Can you create a chart showing sales trends over time?"
      expected_response = ""
    }
  ]
  deletion_protection = false
}

output "is_evaluation_uuid_set" {
  value = (
    lightdash_project_agent_evaluations.test.evaluation_uuid != null
    && length(lightdash_project_agent_evaluations.test.evaluation_uuid) > 0
  )
}
