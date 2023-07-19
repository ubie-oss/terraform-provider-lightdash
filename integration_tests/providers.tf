terraform {
  required_version = "1.1.0"

  required_providers {
    lightdash = {
      source = "github.com/ubie-oss/lightdash"
    }
  }
}

variable "personal_access_token" {
  description = "Lightdash personal access token"
  type        = string
  sensitive   = true
  default     = "00c8e75be71505c3f4c1becc40890508"
}

provider "lightdash" {
  host  = "https://app.lightdash.cloud"
  token = var.personal_access_token
}
