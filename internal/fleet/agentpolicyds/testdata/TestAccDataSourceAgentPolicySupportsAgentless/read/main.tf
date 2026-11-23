provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name               = var.policy_name
  namespace          = "default"
  description        = "Test Agent Policy with supports_agentless"
  monitor_logs       = true
  monitor_metrics    = false
  skip_destroy       = false
  supports_agentless = true
}

data "elasticstack_fleet_agent_policy" "test" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
