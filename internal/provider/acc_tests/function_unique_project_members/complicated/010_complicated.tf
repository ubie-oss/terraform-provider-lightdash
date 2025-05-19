locals {
  unique_project_members = provider::lightdash::unique_project_members(
    ["admin1", "admin1", "admin2"],
    ["admin1", "dev1", "dev1", "dev2"],
    ["admin1", "editor1", "editor1", "editor1", "editor2"],
    ["admin1", "interactive_viewer1", "interactive_viewer1", "interactive_viewer1", "interactive_viewer2"],
    ["admin1", "viewer1", "viewer1", "viewer1", "viewer2"]
  )
}

output "unique_project_members_admins_json" {
  value = jsonencode(local.unique_project_members.admins)
}

output "unique_project_members_developers_json" {
  value = jsonencode(local.unique_project_members.developers)
}

output "unique_project_members_editors_json" {
  value = jsonencode(local.unique_project_members.editors)
}

output "unique_project_members_interactive_viewers_json" {
  value = jsonencode(local.unique_project_members.interactive_viewers)
}

output "unique_project_members_viewers_json" {
  value = jsonencode(local.unique_project_members.viewers)
}
