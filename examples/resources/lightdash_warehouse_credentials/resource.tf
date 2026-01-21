resource "lightdash_warehouse_credentials" "bigquery_prod" {
  organization_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name              = "BigQuery Production"
  description       = "Production BigQuery warehouse credentials"
  warehouse_type    = "bigquery"

  # BigQuery specific configuration
  project          = "my-gcp-project-id"
  dataset          = "my_dataset"
  keyfile_contents = file("${path.module}/service-account-key.json")

  # Optional: BigQuery performance settings
  location               = "US"
  timeout_seconds        = 300
  maximum_bytes_billed   = 1000000000
  priority               = "INTERACTIVE"
  retries                = 3
  start_of_week          = 1
}

# Minimal configuration example
resource "lightdash_warehouse_credentials" "bigquery_minimal" {
  organization_uuid = "xxxxxxxx-xxxxxxxxxx-xxxxxxxxx"
  name              = "BigQuery Minimal"
  warehouse_type    = "bigquery"
  project           = "my-gcp-project-id"
  keyfile_contents  = file("${path.module}/service-account-key.json")
}
