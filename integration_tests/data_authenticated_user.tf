data "lightdash_authenticated_user" "test" {
}

output "lightdash_authenticated_user_test" {
  value = data.lightdash_authenticated_user.test
}
