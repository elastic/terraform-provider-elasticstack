provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Output ${var.policy_name}"
  output_id            = "${var.policy_name}-output"
  type                 = "elasticsearch"
  hosts                = ["https://elasticsearch:9200"]
  default_integrations = false
  default_monitoring   = false
}

resource "elasticstack_fleet_agent_download_source" "test_download_source" {
  name = "Download Source ${var.policy_name}"
  host = "https://artifacts.elastic.co/downloads/elastic-agent-${var.policy_name}"
}

resource "elasticstack_fleet_server_host" "test_host" {
  name  = "Server Host ${var.policy_name}"
  hosts = ["https://fleet-server-${var.policy_name}:8220"]
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "Policy ${var.policy_name}"
  namespace       = "default"
  monitor_logs    = false
  monitor_metrics = false
}
