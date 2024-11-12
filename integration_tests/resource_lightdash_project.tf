# Copyright 2023 Ubie, inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

resource "lightdash_project" "test1" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "zzz_test_project_01_snowflake"
  type = "DEFAULT"
  dbt_connection_repository = "ubie-oss/terraform-provider-lightdash"
  snowflake_warehouse_connection_type = "snowflake"
  snowflake_warehouse_connection_account = "abc-123.eu-west-1"
  snowflake_warehouse_connection_role = "BI_ROLE"
  snowflake_warehouse_connection_database = "DB"
  snowflake_warehouse_connection_warehouse = "TEST_WH"
}

resource "lightdash_project" "test2" {
  organization_uuid = data.lightdash_organization.test.organization_uuid
  name              = "zzz_test_project_02_databricks"
  type = "DEVELOPMENT"
  dbt_connection_repository = "ubie-oss/terraform-provider-lightdash"
  warehouse_connection_type = "databricks"
  databricks_connection_server_host_name = "host-name-for-databricks.com"
  databricks_connection_http_path = "sql/warehouse"
  databricks_connection_personal_access_token = "abcdefg123"
  databricks_connection_catalog = "PROD"
}

output "lightdash_project__test1" {
  value = lightdash_project.test1
}

output "lightdash_project__test2" {
  value = lightdash_project.test2
}
