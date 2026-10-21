data "lightdash_organization" "example" {
}

data "lightdash_project" "example" {
  project_uuid = var.project_uuid
}

resource "lightdash_project_ai_provider_credential" "example" {
  organization_uuid   = data.lightdash_organization.example.organization_uuid
  project_uuid        = data.lightdash_project.example.project_uuid
  credential_uuid     = var.credential_uuid
  deletion_protection = true
}
