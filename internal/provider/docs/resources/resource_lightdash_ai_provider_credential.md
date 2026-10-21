Manages a named Amazon Bedrock BYOK credential at the organization level (the same credentials managed under organization AI provider settings in Lightdash).

Requires an organization admin personal access token. The `api_key` is never returned by the Lightdash API after creation; Terraform keeps the value you supplied in state until you change it. Import does not populate `api_key`—provide a new key and apply to rotate.

Setting `is_default` to `true` makes this credential the organization default via the Lightdash API. Lightdash only supports designating a default credential, not clearing the default through the API; set another credential as default instead.

If your organization still has a legacy Bedrock configuration, adopt it once with the Lightdash API (`POST /api/v1/ai/provider-credentials/adopt-legacy`) before managing credentials with Terraform.

`deletion_protection` is required. When set to `true`, Terraform will not destroy the resource. Imported resources default to `deletion_protection = true`.

Destroy is blocked while the credential is the organization default. Set `is_default = true` on another `lightdash_ai_provider_credential`, apply, then destroy this resource.

Self-hosted instances must expose the AI provider credentials API (same as Lightdash Cloud). Instance-wide `BEDROCK_*` environment variables are not managed by this resource.
