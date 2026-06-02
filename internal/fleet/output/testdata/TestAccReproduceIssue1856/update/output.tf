provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "repro_1856" {
  output_id            = var.output_id
  name                 = "Issue 1856 Repro (updated)"
  type                 = "kafka"
  default_integrations = false
  default_monitoring   = false
  hosts                = ["kafka:9092"]

  kafka = {
    auth_type       = "none"
    connection_type = "plaintext"
    topic           = "test-topic-updated"
    partition       = "round_robin"
  }
}
