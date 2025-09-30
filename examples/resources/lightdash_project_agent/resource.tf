resource "lightdash_project_agent" "test" {
  organization_uuid = "xxxx-xxxx-xxxx"
  project_uuid      = "xxxx-xxxx-xxxx"
  name              = "Test Agent"
  instruction       = "You are a helpful AI assistant for data analysis."

  tags               = ["test", "terraform"]
  enable_data_access = true
  group_access       = ["xxxx-xxxx-xxxx"]
  user_access        = ["xxxx-xxxx-xxxx"]

  # If you want to manually change the instruction on the web UI of LIghtdash,
  # you can ignore the changes to the instruction.
  # lifecycle {
  #   ignore_changes = [
  #     instruction
  #   ]
  # }
}
