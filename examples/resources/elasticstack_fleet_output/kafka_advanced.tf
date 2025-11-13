provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Advanced Kafka Fleet Output with SSL authentication
resource "elasticstack_fleet_output" "kafka_advanced" {
  name                 = "Advanced Kafka Output"
  output_id            = "kafka-advanced-output"
  type                 = "kafka"
  default_integrations = false
  default_monitoring   = false

  hosts = [
    "kafka1:9092",
    "kafka2:9092",
    "kafka3:9092"
  ]

  # Advanced Kafka configuration
  kafka = {
    auth_type      = "ssl"
    topic          = "elastic-logs"
    partition      = "round_robin"
    compression    = "snappy"
    required_acks  = -1
    broker_timeout = 10
    timeout        = 30
    version        = "2.6.0"
    client_id      = "elastic-beats-client"

    # Custom headers for message metadata
    headers = [
      {
        key   = "datacenter"
        value = "us-west-1"
      },
      {
        key   = "service"
        value = "beats"
      },
      {
        key   = "environment"
        value = "production"
      }
    ]

    # Hash-based partitioning
    hash = {
      hash   = "host.name"
      random = false
    }

    # SASL configuration
    sasl = {
      mechanism = "SCRAM-SHA-256"
    }
  }

  # SSL configuration (reusing common SSL block)
  ssl = {
    certificate_authorities = [
      file("${path.module}/ca.crt")
    ]
    certificate = file("${path.module}/client.crt")
    key         = file("${path.module}/client.key")
  }

  # Additional YAML configuration for advanced settings
  config_yaml = yamlencode({
    "ssl.verification_mode"   = "full"
    "ssl.supported_protocols" = ["TLSv1.2", "TLSv1.3"]
    "max.message.bytes"       = 1000000
  })
}

# Example showing round-robin partitioning with event grouping
resource "elasticstack_fleet_output" "kafka_round_robin" {
  name                 = "Kafka Round Robin Output"
  output_id            = "kafka-round-robin-output"
  type                 = "kafka"
  default_integrations = false
  default_monitoring   = false

  hosts = ["kafka:9092"]

  kafka = {
    auth_type   = "none"
    topic       = "elastic-metrics"
    partition   = "round_robin"
    compression = "lz4"

    round_robin = [
      {
        group_events = 100
      }
    ]
  }
}
