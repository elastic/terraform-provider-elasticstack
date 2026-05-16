provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Kafka Output ${var.policy_name}"
  output_id            = "${var.policy_name}-kafka-output"
  type                 = "kafka"
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "kafka:9092"
  ]

  kafka = {
    auth_type       = "none"
    connection_type = "plaintext"
    topic           = "beats"
    partition       = "round_robin"

    round_robin = {
      group_events = 1
    }
  }
}
