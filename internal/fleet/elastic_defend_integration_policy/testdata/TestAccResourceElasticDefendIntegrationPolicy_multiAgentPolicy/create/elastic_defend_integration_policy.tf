variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "one" {
  name      = "${var.policy_name}-agent-policy-one"
  namespace = "default"
}

resource "elasticstack_fleet_agent_policy" "two" {
  name      = "${var.policy_name}-agent-policy-two"
  namespace = "default"
}

resource "elasticstack_fleet_elastic_defend_integration_policy" "test" {
  name                = var.policy_name
  namespace           = "default"
  agent_policy_ids    = [elasticstack_fleet_agent_policy.one.policy_id, elasticstack_fleet_agent_policy.two.policy_id]
  enabled             = true
  integration_version = "8.15.0"
  preset              = "EDRComplete"

  policy = {
    windows = {
      malware = {
        mode = "prevent"
      }
    }
    mac = {
      malware = {
        mode = "prevent"
      }
    }
    linux = {
      malware = {
        mode = "detect"
      }
    }
  }
}
