Manages an OAuth 2.0 application (client) at the organization level. These are the same clients shown under **Organization settings → OAuth applications** in the Lightdash UI.

Requires an organization admin personal access token. The `client_secret` is returned only once at creation and is not refreshed on read. Import does not populate `client_secret`; create a new application to rotate credentials.

`redirect_uris` is a full replacement list on update (the Lightdash API PATCH replaces the entire array).

`deletion_protection` is required. When set to `true`, Terraform will not destroy the resource. Imported resources default to `deletion_protection = true`.
