variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_integration" "pubsub" {
  name = "gcp_pubsub"
}

resource "elasticstack_fleet_integration" "pubsub" {
  name    = "gcp_pubsub"
  version = data.elasticstack_fleet_integration.pubsub.version
  force   = true
}

resource "elasticstack_fleet_agent_policy" "example" {
  name            = "${var.policy_name} Agent Policy"
  namespace       = "default"
  description     = "GCP PubSub Test Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
}

resource "elasticstack_fleet_integration_policy" "pubsub" {
  name                = var.policy_name
  namespace           = "default"
  integration_name    = elasticstack_fleet_integration.pubsub.name
  integration_version = elasticstack_fleet_integration.pubsub.version
  agent_policy_id     = elasticstack_fleet_agent_policy.example.policy_id

  inputs = {
    "gcp-gcp-pubsub" = {
      enabled = true
      streams = {
        "gcp_pubsub.gcp" = {
          enabled = true
          vars = jsonencode({
            project_id        = "my-project"
            topic             = "my-topic"
            subscription_name = "my-sub"
            tags              = ["forwarded"]
          })
        }
      }
    }
  }
}
