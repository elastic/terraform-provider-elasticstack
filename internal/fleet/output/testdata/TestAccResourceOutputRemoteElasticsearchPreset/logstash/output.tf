provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Remote Elasticsearch Preset Output ${var.policy_name}"
  output_id            = "${var.policy_name}-remote-elasticsearch-preset-output"
  type                 = "logstash"
  default_integrations = false
  default_monitoring   = false

  hosts = [
    "logstash:5044",
  ]
}
