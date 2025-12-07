provider "elasticstack" {
  elasticsearch {}
  kibana {}
}
variable "policy_name" {
  type = string
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
}
