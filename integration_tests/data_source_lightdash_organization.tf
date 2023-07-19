data "lightdash_organization" "test" {}

output "lightdash_organization_test" {
  value = data.lightdash_organization.test
}
