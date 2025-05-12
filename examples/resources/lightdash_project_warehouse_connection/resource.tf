# resource "lightdash_project_warehouse_connection" "example" {
#   project_id = "xxxx-xxx-xxx-xxx"

#   bigquery {
#     google_project_id           = "xxxx-xxx-xxx-xxx"
#     location                    = "xxxx"
#     execution_google_project_id = null
#     query_timeout               = 300
#     priority                    = "interactive" # or "batch"
#     retries                     = 3
#     key_file                    = hoge
#     maximum_bytes_billed        = 1000000000
#     state_of_week               = null
#     project                     = "xxxx-xxx-xxx-xxx"
#     timeoutSeconds              = 300
#     keyfileContents             = file("xxxx-xxx-xxx-xxx")
#     maximumBytesBilled          = 1000000000
#     startOfWeek                 = null
#     executionProject            = "xxxx-xxx-xxx-xxx"
#   }
# }
