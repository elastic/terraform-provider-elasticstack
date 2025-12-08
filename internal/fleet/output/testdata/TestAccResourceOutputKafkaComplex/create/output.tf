provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name      = "Complex Kafka Output ${var.policy_name}"
  output_id = "${var.policy_name}-kafka-complex-output"
  type      = "kafka"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "kafka1:9092",
    "kafka2:9092",
    "kafka3:9092"
  ]

  # Complex Kafka configuration showcasing all options
  kafka = {
    auth_type       = "none"
    topic           = "complex-topic"
    partition       = "hash"
    compression     = "lz4"
    connection_type = "encryption"
    required_acks   = 0
    broker_timeout  = 10
    timeout         = 30
    version         = "2.6.0"

    headers = [
      {
        key   = "datacenter"
        value = "us-west-1"
      },
      {
        key   = "service"
        value = "beats"
      }
    ]

    hash = {
      hash   = "event.hash"
      random = false
    }

    sasl = {
      mechanism = "SCRAM-SHA-256"
    }
  }
}
