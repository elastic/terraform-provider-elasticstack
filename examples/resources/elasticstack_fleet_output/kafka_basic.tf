provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Basic Kafka Fleet Output
resource "elasticstack_fleet_output" "kafka_basic" {
  name                 = "Basic Kafka Output"
  output_id            = "kafka-basic-output"
  type                 = "kafka"
  default_integrations = false
  default_monitoring   = false

  hosts = [
    "kafka:9092"
  ]

  # Basic Kafka configuration
  kafka = {
    auth_type     = "user_pass"
    username      = "kafka_user"
    password      = "kafka_password"
    topic         = "elastic-beats"
    partition     = "hash"
    compression   = "gzip"
    required_acks = 1

    headers = [
      {
        key   = "environment"
        value = "production"
      }
    ]
  }
}
