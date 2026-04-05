variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test" {
  name      = "${var.policy_name}-agent-policy"
  namespace = "default"
}

resource "elasticstack_fleet_elastic_defend_integration_policy" "test" {
  name                = var.policy_name
  namespace           = "default"
  agent_policy_id     = elasticstack_fleet_agent_policy.test.policy_id
  integration_version = "8.14.0"
  preset              = "NGAv1"

  policy = {
    windows = {
      events = {
        process = true
        network = true
        file    = true
      }
      malware = {
        mode = "prevent"
      }
    }
    mac = {
      events = {
        process = true
      }
    }
    linux = {
      events = {
        process = true
        network = true
      }
    }
  }
}
