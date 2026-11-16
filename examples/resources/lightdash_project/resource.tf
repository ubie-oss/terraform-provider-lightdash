resource "lightdash_project" "example" {
  name                = "example-project"
  type                = "DEFAULT"
  dbt_version         = "latest"
  deletion_protection = true
}
