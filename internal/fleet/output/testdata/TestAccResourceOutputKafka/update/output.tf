provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name      = "Updated Kafka Output ${var.policy_name}"
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

  # Updated Kafka-specific configuration
  kafka = {
    auth_type       = "none"
    topic           = "logs"
    partition       = "round_robin"
    compression     = "snappy"
    connection_type = "encryption"
    required_acks   = -1
  }
}
