# Space can be impported by specifing the resource identifier.
terraform import lightdash_group.example "organizations/${organizatio_uuid}/groups/${group_uuid}"

terraform import -var-file="testing.tfvars" \
	lightdash_group.test2 "organizations/089a18c4-667e-41cb-9d10-b088461ac941/groups/489aefd3-01b5-481f-b4a4-b7a134a18be0"
