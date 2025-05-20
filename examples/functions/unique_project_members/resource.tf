locals {
  unique_project_members = provider::lightdash::unique_project_members(
    ["admin1", "admin2", "admin2"],
    [
      "admin1", "developer1",
      "developer2", "developer2",
    ],
    [
      "admin1", "developer1", "editor1",
      "editor2", "editor2",
    ],
    [
      "admin1", "developer1", "editor1",
      "interactive_viewer1", "interactive_viewer2", "interactive_viewer2",
    ],
    [
      "admin1", "developer1", "editor1", "interactive_viewer1",
      "viewer1", "viewer2", "viewer2",
    ]
  )
}

output "unique_project_members_admins" {
  description = "List of admin member UUIDs. This should be '[admin1, admin2]'"
  value       = local.unique_project_members.admins
}

output "unique_project_members_developers" {
  description = "List of developer member UUIDs. This should be '[developer1, developer2]'"
  value       = local.unique_project_members.developers
}

output "unique_project_members_editors" {
  description = "List of editor member UUIDs. This should be '[editor1, editor2]'"
  value       = local.unique_project_members.editors
}

output "unique_project_members_interactive_viewers" {
  description = "List of interactive viewer member UUIDs. This should be '[interactive_viewer1, interactive_viewer2]'"
  value       = local.unique_project_members.interactive_viewers
}

output "unique_project_members_viewers" {
  description = "List of viewer member UUIDs. This should be '[viewer1, viewer2]'"
  value       = local.unique_project_members.viewers
}
