resource "lightdash_project_role_member" "test" {
  project_uuid = "xxx-xxx-xxxx"
  user_uuid    = "xxxxxx-xxxxxxxxx-xxxxx"
  role         = "editor"
}
