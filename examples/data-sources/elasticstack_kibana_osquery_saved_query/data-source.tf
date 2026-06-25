provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_saved_query" "example" {
  saved_query_id = "list_processes"
  query          = "SELECT pid, name FROM processes LIMIT 10;"
  interval       = 3600
}

# Read a user-managed query by ID (created above or outside Terraform).
# A common use is looking up saved query IDs referenced by Security detection
# rule response actions (response_actions[].params.saved_query_id).
data "elasticstack_kibana_osquery_saved_query" "managed" {
  saved_query_id = elasticstack_kibana_osquery_saved_query.example.saved_query_id
}

# Read a prebuilt query from the osquery_manager integration.
# Prebuilt IDs vary by integration version; replace with an ID from your deployment
# (for example processes_elastic from the Kibana API client examples).
data "elasticstack_kibana_osquery_saved_query" "prebuilt" {
  saved_query_id = "processes_elastic"
}

output "managed_query" {
  value = data.elasticstack_kibana_osquery_saved_query.managed.query
}

output "prebuilt" {
  value = data.elasticstack_kibana_osquery_saved_query.prebuilt.prebuilt
}
