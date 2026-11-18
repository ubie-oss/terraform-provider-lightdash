data "lightdash_user_attribute" "region" {
  name = "region"
}

output "region_user_attribute_uuid" {
  value = data.lightdash_user_attribute.region.user_attribute_uuid
}
