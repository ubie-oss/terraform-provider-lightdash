data "lightdash_organization" "example" {
}

resource "lightdash_ai_provider_credential" "bedrock" {
  organization_uuid   = data.lightdash_organization.example.organization_uuid
  ai_provider         = "bedrock"
  label               = "production-bedrock"
  region              = "us-east-1"
  allowed_models      = ["claude-sonnet-4-5"]
  inference_geography = "us"
  api_key             = var.bedrock_api_key
  is_default          = true
  deletion_protection = true
}
