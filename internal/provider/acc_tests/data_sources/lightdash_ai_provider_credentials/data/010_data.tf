data "lightdash_organization" "test" {
}

data "lightdash_ai_provider_credentials" "test" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
}
