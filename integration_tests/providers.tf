terraform {
  required_version = "1.1.0"

  required_providers {
    lightdash = {
      source = "github.com/ubie-oss/lightdash"
    }
  }
}

provider "lightdash" {
  host  = var.lightdash_url
  token = var.personal_access_token
}
