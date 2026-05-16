provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                   = "Elasticsearch Output ${var.policy_name}"
  output_id              = "${var.policy_name}-elasticsearch-output"
  type                   = "elasticsearch"
  default_integrations   = false
  default_monitoring     = false
  ca_sha256              = "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
  ca_trusted_fingerprint = "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567891"
  hosts = [
    "https://elasticsearch:9200"
  ]
}
