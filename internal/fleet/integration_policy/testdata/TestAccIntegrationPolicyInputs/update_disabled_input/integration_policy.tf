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
  description         = "Kafka Integration Policy - Updated"

  inputs = {
    "kafka-logfile" = {
      enabled = false 
      # Should not need to specify streams for disabled input
    }
    "kafka-kafka/metrics" = {
      enabled = true
      vars = jsonencode({
        hosts   = ["localhost:9092"]
        period  = "10s"
        "ssl.certificate_authorities"     =[]
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