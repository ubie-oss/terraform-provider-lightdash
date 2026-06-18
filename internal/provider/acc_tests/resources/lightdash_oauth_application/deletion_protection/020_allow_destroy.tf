data "lightdash_organization" "test" {
}

resource "lightdash_oauth_application" "test" {
  organization_uuid   = data.lightdash_organization.test.organization_uuid
  client_name         = "test (Acceptance Test - oauth deletion_protection)"
  redirect_uris       = ["https://example.com/oauth/callback"]
  deletion_protection = false
}
