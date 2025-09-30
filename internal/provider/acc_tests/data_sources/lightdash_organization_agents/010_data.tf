data "lightdash_organization" "test" {}

data "lightdash_organization_agents" "test" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
}
