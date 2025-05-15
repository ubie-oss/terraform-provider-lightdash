provider "lightdash" {
  host  = env("LIGHTDASH_URL")
  token = env("LIGHTDASH_API_KEY")
}

data "lightdash_project" "test" {
  project_uuid = env("LIGHTDASH_PROJECT")
}
