locals {
  unique_project_members = provider::lightdash::unique_project_members(
    ["admin1"],
    ["dev1"],
    ["editor1"],
    ["interactive_viewer1"],
    ["viewer1"]
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
