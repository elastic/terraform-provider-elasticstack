provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = var.space_id
  name        = var.space_name
  description = "Test space for fleet output space import test"
}

resource "elasticstack_fleet_output" "test_output" {
  name      = "Elasticsearch Output ${var.policy_name}"
  output_id = "${var.policy_name}-elasticsearch-output"
  type      = "elasticsearch"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "https://elasticsearch:9200"
  ]
  space_ids = [var.space_id]

  depends_on = [elasticstack_kibana_space.test_space]
}
