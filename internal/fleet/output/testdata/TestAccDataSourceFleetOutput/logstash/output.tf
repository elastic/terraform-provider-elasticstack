provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_fleet_output" "logstash" {
  name      = "Logstash Output ${var.policy_name}"
  output_id = "${var.policy_name}-logstash-output"
  type      = "logstash"
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

data "elasticstack_fleet_output" "logstash" {
  output_id = elasticstack_fleet_output.logstash.output_id
}
