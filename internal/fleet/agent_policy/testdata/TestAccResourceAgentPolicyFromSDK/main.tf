variable "policy_name" {
  type        = string
  description = "Name for the agent policy"
}

variable "skip_destroy" {
  type        = bool
  description = "Whether to skip destruction of the policy"
  default     = false
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy"
  monitor_logs    = true
  monitor_metrics = false
  skip_destroy    = var.skip_destroy
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}