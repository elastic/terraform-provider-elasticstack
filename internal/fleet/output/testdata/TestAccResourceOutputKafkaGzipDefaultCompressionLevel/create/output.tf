provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Kafka Gzip Default Level ${var.policy_name}"
  output_id            = "${var.policy_name}-kafka-gzip-default"
  type                 = "kafka"
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "kafka:9092"
  ]

  kafka = {
    auth_type       = "none"
    topic           = "beats"
    partition       = "hash"
    compression     = "gzip"
    connection_type = "plaintext"
    required_acks   = 1
  }
}
