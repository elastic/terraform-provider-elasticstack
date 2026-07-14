provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Kafka User Pass No SASL ${var.policy_name}"
  output_id            = "${var.policy_name}-kafka-user-pass-no-sasl"
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
  }
}
