variable "policy_name" {
  type = string
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "TestAccIntegrationPolicyInputs Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

data "elasticstack_fleet_integration" "test" {
  name = "kafka"
}

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name                = var.policy_name
  namespace           = "default"
  agent_policy_id     = elasticstack_fleet_agent_policy.test_policy.id
  integration_name    = "kafka"
  integration_version = data.elasticstack_fleet_integration.test.version
  description         = "Kafka Integration Policy"

  inputs = {
    "kafka-logfile" = {
      enabled = true
      streams = {
        "kafka.log" = {
          enabled = true
          vars = jsonencode({
            "kafka_home" = "/opt/kafka*",
            "paths" = [
              "/logs/controller.log*",
              "/logs/server.log*",
              "/logs/state-change.log*",
              "/logs/kafka-*.log*"
            ],
            "tags" = [
              "kafka-log"
            ],
            "preserve_original_event" = false
          })
        }
      }
    }
    "kafka-kafka/metrics" = {
      enabled = true
      vars = jsonencode({
        hosts                         = ["localhost:9092"]
        period                        = "10s"
        "ssl.certificate_authorities" = []
      })
      streams = {
        "kafka.broker" = {
          enabled = true
          vars = jsonencode({
            "jolokia_hosts" = ["localhost:8778"]
          })
        }
        "kafka.consumergroup" = {
          enabled = true
          vars = jsonencode({
            "topics" = []
          })
        }
        "kafka.partition" = {
          enabled = false
          vars = jsonencode({
            "topics" = []
          })
        }
      }
    }
  }
}