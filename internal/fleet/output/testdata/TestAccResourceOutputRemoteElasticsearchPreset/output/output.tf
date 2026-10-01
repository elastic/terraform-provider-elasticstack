provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Remote Elasticsearch Preset Output ${var.policy_name}"
  output_id            = "${var.policy_name}-remote-elasticsearch-preset-output"
  type                 = "remote_elasticsearch"
  service_token        = var.service_token
  preset               = var.preset
  default_integrations = false
  default_monitoring   = false

  hosts = [
    "https://elasticsearch:9200",
  ]
}
