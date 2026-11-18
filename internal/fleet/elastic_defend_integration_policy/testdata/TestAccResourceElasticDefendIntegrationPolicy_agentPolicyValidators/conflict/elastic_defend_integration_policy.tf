variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_elastic_defend_integration_policy" "test" {
  name                = var.policy_name
  namespace           = "default"
  agent_policy_id     = "dummy-agent-policy"
  agent_policy_ids    = ["dummy-agent-policy"]
  enabled             = true
  integration_version = "8.14.0"
  preset              = "EDRComplete"

  policy = {
    windows = {
      malware = {
        mode = "prevent"
      }
    }
  }
}
