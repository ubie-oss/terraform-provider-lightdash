resource "lightdash_organization_role_member" "test" {
  organization_uuid = "xxxxx-xxxxxx-xxxx"
  user_uuid         = "xxxx-xxx-xxx"
  role              = "editor"
}
