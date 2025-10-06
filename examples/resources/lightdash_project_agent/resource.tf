resource "lightdash_project_agent" "test" {
  organization_uuid = "xxxx-xxxx-xxxx"
  project_uuid      = "xxxx-xxxx-xxxx"
  name              = "Test Agent"
  instruction       = "You are a helpful AI assistant for data analysis."

  tags = ["test", "terraform"]

  enable_data_access      = true
  enable_self_improvement = true

  group_access = ["xxxx-xxxx-xxxx"]
  user_access  = ["xxxx-xxxx-xxxx"]

  deletion_protection = true

  integrations = [
    {
      type       = "slack"
      channel_id = "<YOUR_SLACK_CHANNEL_ID>"
    }
  ]

  # If you want to manually change the instruction on the web UI of LIghtdash,
  # you can ignore the changes to the instruction.
  # lifecycle {
  #   ignore_changes = [
  #     instruction
  #   ]
  # }
}
