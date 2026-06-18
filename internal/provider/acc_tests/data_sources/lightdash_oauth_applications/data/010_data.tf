data "lightdash_organization" "test" {
}

resource "lightdash_oauth_application" "acc_test" {
  organization_uuid   = data.lightdash_organization.test.organization_uuid
  client_name         = "test (Acceptance Test - oauth data sources)"
  redirect_uris       = ["https://example.com/oauth/callback"]
  deletion_protection = false
}

data "lightdash_oauth_application" "by_id" {
  client_id = lightdash_oauth_application.acc_test.client_id

  depends_on = [lightdash_oauth_application.acc_test]
}

data "lightdash_oauth_applications" "all" {
  organization_uuid = data.lightdash_organization.test.organization_uuid

  depends_on = [lightdash_oauth_application.acc_test]
}
