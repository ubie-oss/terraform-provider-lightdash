resource "lightdash_project_scheduler_settings" "example" {
  organization_uuid = "org-1234567890"
  project_uuid      = "proj-1234567890"

  scheduler_timezone = "Asia/Tokyo"
}
