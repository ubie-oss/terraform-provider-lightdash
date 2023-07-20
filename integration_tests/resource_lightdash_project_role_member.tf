resource "lightdash_project_role_member" "test" {
  project_uuid = "866d0073-b79e-4100-a8ff-7f371289609b"
  user_uuid    = "ed4b34e3-390a-4456-9188-f8947f8e600a"
  role         = "editor"
}

output "lightdash_project_role_member__test" {
  value     = lightdash_project_role_member.test
  sensitive = true
}
