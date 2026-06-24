provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Read a prebuilt query shipped with the osquery_manager integration.
# Prebuilt queries cannot be managed by the resource; use this data source instead.
data "elasticstack_kibana_osquery_saved_query" "prebuilt" {
  saved_query_id = "list_all_processes"
}

# Read a user-managed query created outside Terraform (or by the resource).
data "elasticstack_kibana_osquery_saved_query" "external" {
  saved_query_id = "list_processes"
  space_id       = "default"
}

output "prebuilt_query" {
  value = data.elasticstack_kibana_osquery_saved_query.prebuilt.query
}

output "prebuilt" {
  value = data.elasticstack_kibana_osquery_saved_query.prebuilt.prebuilt
}
