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
    auth_type = "user_pass"
    topic     = "beats"
    username  = "testuser"
    password  = "testpass"

    sasl = {
      mechanism = "PLAIN"
    }
  }
}
