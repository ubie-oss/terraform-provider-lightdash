resource "lightdash_group" "test_group1" {
  organization_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name              = "test_group1"
}


resource "lightdash_group" "test_group2" {
  organization_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name              = "test_group2"

  members = [
    {
      user_uuid = "xxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxx"
    },
    {
      user_uuid = "yyyyyyyyyyy-yyyyyyyyyyyy-yyyyyyyyyy"
    },
  ]
}
