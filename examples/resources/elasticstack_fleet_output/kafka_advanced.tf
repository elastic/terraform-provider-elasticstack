provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Placeholder PEM material for illustration only — replace with real certificates in production.
locals {
  example_ca          = <<-EOT
    -----BEGIN CERTIFICATE-----
    MIIBkTCB+wIJAKHHCgV4Jh0FMA0GCSqGGSIb3DQEBCwUAMBExCzAJBgNVBAYTAlVT
    -----END CERTIFICATE-----
  EOT
  example_client_cert = <<-EOT
    -----BEGIN CERTIFICATE-----
    MIIBkTCB+wIJAKHHCgV4Jh0FMA0GCSqGGSIb3DQEBCwUAMBExCzAJBgNVBAYTAlVT
    -----END CERTIFICATE-----
  EOT
  example_client_key  = <<-EOT
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpAIBAAKCAQEA0
    -----END RSA PRIVATE KEY-----
  EOT
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

  kafka = {
    auth_type      = "ssl"
    topic          = "elastic-logs"
    partition      = "round_robin"
    compression    = "snappy"
    required_acks  = -1
    broker_timeout = 10.0
    timeout        = 30.0
    version        = "2.6.0"
    client_id      = "elastic-beats-client"

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
  }

  ssl = {
    certificate_authorities = [trimspace(local.example_ca)]
    certificate             = trimspace(local.example_client_cert)
    key                     = trimspace(local.example_client_key)
  }

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

    round_robin = {
      group_events = 100
    }
  }
}
