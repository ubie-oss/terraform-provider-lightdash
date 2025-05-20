locals {
  all_members_emails = [
    for member in data.lightdash_organization_members.all_members.members : member.email
  ]
}

data "lightdash_organization_members" "all_members" {
}

// Query all members
data "lightdash_organization_members_by_emails" "all_members" {
  emails = local.all_members_emails
}

// Query one member
data "lightdash_organization_members_by_emails" "one_member" {
  emails = [
    local.all_members_emails[0],
  ]
}
