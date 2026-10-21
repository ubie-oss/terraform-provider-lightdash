data "lightdash_organization" "test" {
}

resource "lightdash_ai_provider_credential" "bedrock" {
  organization_uuid   = data.lightdash_organization.test.organization_uuid
  ai_provider         = "bedrock"
  label               = "test (Acceptance Test - project ai binding)"
  region              = "us-east-1"
  allowed_models      = ["claude-sonnet-4-5"]
  api_key             = "terraform-acc-dummy-bedrock-key-not-for-inference"
  deletion_protection = false
}

resource "lightdash_project_ai_provider_credential" "test" {
  organization_uuid   = data.lightdash_organization.test.organization_uuid
  project_uuid        = data.lightdash_project.test.project_uuid
  credential_uuid     = lightdash_ai_provider_credential.bedrock.credential_uuid
  deletion_protection = false
}
