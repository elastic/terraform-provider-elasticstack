provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_fleet_output" "test_output" {
  name      = "Logstash Output ${var.policy_name}"
  type      = "logstash"
  output_id = "${var.policy_name}-logstash-output"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "logstash:5044"
  ]
  ssl = {
    certificate_authorities = ["placeholder"]
    certificate             = "placeholder"
    key                     = "placeholder"
  }
}
