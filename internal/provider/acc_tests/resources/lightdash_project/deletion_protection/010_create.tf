resource "lightdash_project" "test" {
  name                = "tf-acc-lightdash-project-protected"
  type                = "DEFAULT"
  dbt_version         = "latest"
  deletion_protection = true
}
