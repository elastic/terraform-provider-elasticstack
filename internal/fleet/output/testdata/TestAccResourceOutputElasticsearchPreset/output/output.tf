provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Elasticsearch Preset Output ${var.policy_name}"
  output_id            = "${var.policy_name}-elasticsearch-preset-output"
  type                 = "elasticsearch"
  preset               = var.preset
  config_yaml          = var.config_yaml
  default_integrations = false
  default_monitoring   = false

  hosts = [
    var.host,
  ]
}
