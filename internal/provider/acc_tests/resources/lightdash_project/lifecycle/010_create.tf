resource "lightdash_project" "test" {
  name                = "tf-acc-lightdash-project"
  type                = "DEFAULT"
  dbt_version         = "latest"
  deletion_protection = false
}
