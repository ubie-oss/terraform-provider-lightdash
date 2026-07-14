resource "lightdash_project_upstream" "example" {
  organization_uuid     = "org-1234567890"
  project_uuid          = "proj-dev-1234567890"
  upstream_project_uuid = "proj-prod-1234567890"
}
