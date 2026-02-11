provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_fleet_output" "elasticsearch" {
  name                   = "Elasticsearch Output ${var.policy_name}"
  output_id              = "${var.policy_name}-elasticsearch-output"
  type                   = "elasticsearch"
  ca_sha256              = "sha256fingerprint"
  ca_trusted_fingerprint = "trustedfingerprint"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "https://elasticsearch:9200"
  ]
}

data "elasticstack_fleet_output" "elasticsearch" {
  output_id = elasticstack_fleet_output.elasticsearch.output_id
}
