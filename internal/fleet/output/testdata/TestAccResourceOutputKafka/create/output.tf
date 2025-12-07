provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name      = "Kafka Output ${var.policy_name}"
  output_id = "${var.policy_name}-kafka-output"
  type      = "kafka"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "kafka:9092"
  ]

  # Kafka-specific configuration
  kafka = {
    auth_type         = "none"
    topic             = "beats"
    partition         = "hash"
    compression       = "gzip"
    compression_level = 6
    connection_type   = "plaintext"
    required_acks     = 1

    headers = [{
      key   = "environment"
      value = "test"
    }]
  }
}
