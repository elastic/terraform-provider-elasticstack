variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

variable "space_id" {
  description = "An additional Kibana space ID used by the update step"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test space for Elastic Defend space_ids"
  description = "Test space for Elastic Defend space_ids acceptance test"
}

resource "elasticstack_fleet_agent_policy" "test" {
  name      = "${var.policy_name}-agent-policy"
  namespace = "default"
  space_ids = ["default"]

  depends_on = [elasticstack_kibana_space.test]
}

resource "elasticstack_fleet_elastic_defend_integration_policy" "test" {
  name                = var.policy_name
  namespace           = "default"
  agent_policy_id     = elasticstack_fleet_agent_policy.test.policy_id
  enabled             = true
  integration_version = "8.14.0"
  preset              = "EDRComplete"
  space_ids           = ["default"]

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
