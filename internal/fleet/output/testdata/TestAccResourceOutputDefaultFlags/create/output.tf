provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Elasticsearch Output ${var.policy_name}"
  output_id            = "${var.policy_name}-elasticsearch-output"
  type                 = "elasticsearch"
  default_integrations = true
  default_monitoring   = true
  hosts = [
    "https://elasticsearch:9200"
  ]
}
