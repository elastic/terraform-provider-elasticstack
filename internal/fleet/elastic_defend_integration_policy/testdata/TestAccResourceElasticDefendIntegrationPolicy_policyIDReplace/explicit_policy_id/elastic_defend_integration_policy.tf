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
  enabled             = true
  integration_version = "8.14.0"
  preset              = "EDRComplete"
  # Explicitly configuring policy_id with a different value than the
  # server-assigned one currently in state must force a replace, since
  # policy_id carries the RequiresReplace plan modifier.
  policy_id = "explicit-policy-id-does-not-match-state"

  policy = {
    windows = {
      malware = {
        mode = "prevent"
      }
    }
  }
}
