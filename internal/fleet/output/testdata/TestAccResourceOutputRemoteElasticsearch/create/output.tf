provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Remote Elasticsearch Output ${var.policy_name}"
  output_id            = "${var.policy_name}-remote-elasticsearch-output"
  type                 = "remote_elasticsearch"
  service_token        = var.service_token
  sync_integrations    = false
  sync_uninstalled_integrations = false
  write_to_logs_streams = false
  default_integrations = false
  default_monitoring   = false

  hosts = [
    "https://elasticsearch:9200",
  ]
}
