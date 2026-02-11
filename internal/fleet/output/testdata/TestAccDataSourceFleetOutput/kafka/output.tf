provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_fleet_output" "kafka" {
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

  kafka = {
    auth_type         = "none"
    topic             = "beats"
    partition         = "hash"
    compression       = "gzip"
    compression_level = 6
    connection_type   = "plaintext"
    required_acks     = 1
    broker_timeout    = 10
    timeout           = 30
    version           = "2.6.0"
    client_id         = "fleet-output-client"
    key               = "event.key"

    headers = [{
      key   = "environment"
      value = "test"
    }]

    hash = {
      hash   = "event.hash"
      random = false
    }

    sasl = {
      mechanism = "SCRAM-SHA-256"
    }
  }
}

data "elasticstack_fleet_output" "kafka" {
  output_id = elasticstack_fleet_output.kafka.output_id
}
