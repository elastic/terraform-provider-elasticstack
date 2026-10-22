variable "policy_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "Policy ${var.policy_name}"
  namespace       = "default"
  monitor_logs    = false
  monitor_metrics = false
}
