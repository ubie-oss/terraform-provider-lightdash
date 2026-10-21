data "lightdash_organization" "test" {
}

resource "lightdash_ai_provider_credential" "test" {
  organization_uuid   = data.lightdash_organization.test.organization_uuid
  ai_provider         = "bedrock"
  label               = "test (Acceptance Test - ai credential import)"
  region              = "us-east-1"
  allowed_models      = ["claude-sonnet-4-5"]
  api_key             = "terraform-acc-dummy-bedrock-key-not-for-inference"
  deletion_protection = false
}
