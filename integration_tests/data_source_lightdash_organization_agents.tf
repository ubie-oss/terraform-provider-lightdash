data "lightdash_organization_agents" "test" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
}

output "lightdash_organization_agents_test" {
  value = data.lightdash_organization_agents.test
}
