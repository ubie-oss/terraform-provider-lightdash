locals {
  normalized_project_members = provider::lightdash::normalize_project_members(
    ["admin1", "admin1", "admin2"],
    ["admin1", "dev1", "dev1", "dev2"],
    ["admin1", "editor1", "editor1", "editor1", "editor2"],
    ["admin1", "interactive_viewer1", "interactive_viewer1", "interactive_viewer1", "interactive_viewer2"],
    ["admin1", "viewer1", "viewer1", "viewer1", "viewer2"]
  )
}

output "normalized_project_members_admins_json" {
  value = jsonencode(local.normalized_project_members.admins)
}

output "normalized_project_members_developers_json" {
  value = jsonencode(local.normalized_project_members.developers)
}

output "normalized_project_members_editors_json" {
  value = jsonencode(local.normalized_project_members.editors)
}

output "normalized_project_members_interactive_viewers_json" {
  value = jsonencode(local.normalized_project_members.interactive_viewers)
}

output "normalized_project_members_viewers_json" {
  value = jsonencode(local.normalized_project_members.viewers)
}
