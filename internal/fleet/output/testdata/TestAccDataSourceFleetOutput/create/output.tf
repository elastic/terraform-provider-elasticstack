provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_fleet_output" "elasticsearch" {
  name                   = "Elasticsearch Output ${var.policy_name}"
  output_id              = "${var.policy_name}-elasticsearch-output"
  type                   = "elasticsearch"
  ca_sha256              = "sha256fingerprint"
  ca_trusted_fingerprint = "trustedfingerprint"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "https://elasticsearch:9200"
  ]
}

resource "elasticstack_fleet_output" "logstash" {
  name      = "Logstash Output ${var.policy_name}"
  output_id = "${var.policy_name}-logstash-output"
  type      = "logstash"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "logstash:5044"
  ]
  ssl = {
    certificate_authorities = ["placeholder"]
    certificate             = "placeholder"
    key                     = "placeholder"
  }
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

    headers = [{
      key   = "environment"
      value = "test"
    }]
  }
}

data "elasticstack_fleet_output" "elasticsearch" {
  output_id = elasticstack_fleet_output.elasticsearch.output_id
}

data "elasticstack_fleet_output" "logstash" {
  output_id = elasticstack_fleet_output.logstash.output_id
}

data "elasticstack_fleet_output" "kafka" {
  output_id = elasticstack_fleet_output.kafka.output_id
}
