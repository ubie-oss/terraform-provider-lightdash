variable "lightdash_url" {
  description = "The URL of the Lightdash instance to test."
  type        = string
}

variable "personal_access_token" {
  description = "Lightdash personal access token"
  type        = string
  sensitive   = true
}

variable "test_lightdash_project_uuid" {
  description = "The UUID of the Lightdash project to test."
  type        = string
}

variable "test_organization_admin_user_email" {
  description = "The email of the Lightdash organization admin user to test."
  type        = string
}

variable "test_organization_member_user_email" {
  description = "The email of the Lightdash organization member user to test."
  type        = string
  default     = null
}
