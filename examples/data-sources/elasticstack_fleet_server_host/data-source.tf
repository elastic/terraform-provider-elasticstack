# Excluded from TestAccExamples_planOnly (planOnlySkippedEmbedPaths): the lookups fail
# unless a matching Fleet server host already exists in Kibana/Fleet, and host names and
# IDs vary by stack. Substitute the name or host_id of your Fleet server host before apply.

provider "elasticstack" {
  kibana {}
}

data "elasticstack_fleet_server_host" "by_name" {
  name = "Fleet"
}

data "elasticstack_fleet_server_host" "by_id" {
  host_id = "fleet-default-fleet-server-host"
}

data "elasticstack_fleet_server_host" "default" {
  default = true
}

resource "elasticstack_fleet_agent_policy" "example" {
  name                 = "example"
  namespace            = "default"
  fleet_server_host_id = data.elasticstack_fleet_server_host.by_name.host_id
}
