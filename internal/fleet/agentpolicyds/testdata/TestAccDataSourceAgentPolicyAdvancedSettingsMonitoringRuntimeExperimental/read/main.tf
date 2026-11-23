provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy with monitoring_runtime_experimental"
  monitor_logs    = true
  monitor_metrics = true

  advanced_settings = {
    monitoring_runtime_experimental = "process"
  }
}

data "elasticstack_fleet_agent_policy" "test" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
