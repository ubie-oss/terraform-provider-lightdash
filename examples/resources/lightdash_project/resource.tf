resource "lightdash_project" "test_snowflake" {
  organization_uuid                        = "xxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxx"
  name                                     = "test_project_snowflake"
  type                                     = "DEFAULT"
  dbt_connection_repository                = "xxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxx"
  snowflake_warehouse_connection_type      = "snowflake"
  snowflake_warehouse_connection_account   = "xxxx.xxxxx"
  snowflake_warehouse_connection_role      = "xxxx"
  snowflake_warehouse_connection_database  = "xxx"
  snowflake_warehouse_connection_warehouse = "xxxx"
}

resource "lightdash_project" "test_databricks" {
  organization_uuid                           = "xxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxx"
  name                                        = "test_project_databricks"
  type                                        = "DEVELOPMENT"
  dbt_connection_repository                   = "xxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxx"
  warehouse_connection_type                   = "databricks"
  databricks_connection_server_host_name      = "xxxx.com"
  databricks_connection_http_path             = "xxx/xxxxxx"
  databricks_connection_personal_access_token = "xxxxxxx"
  databricks_connection_catalog               = "xxxx"
}
