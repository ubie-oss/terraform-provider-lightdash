Selects which organization AI provider credential a Lightdash project uses for AI features. Omit `credential_uuid` or set it to null to use the organization default credential.

Requires a token with permission to update project settings. Destroying this resource resets the project to the organization default (`credentialUuid: null` in the Lightdash API) unless `deletion_protection` is enabled.

`deletion_protection` is required. Imported resources default to `deletion_protection = true`.
