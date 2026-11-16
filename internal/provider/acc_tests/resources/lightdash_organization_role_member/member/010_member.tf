variable "org_member_email" {
  type      = string
  sensitive = true
}

data "lightdash_organization" "test" {}

data "lightdash_organization_member" "test" {
  email = var.org_member_email
}

resource "lightdash_organization_role_member" "test" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  user_uuid         = data.lightdash_organization_member.test.user_uuid
  role              = "member"
}
