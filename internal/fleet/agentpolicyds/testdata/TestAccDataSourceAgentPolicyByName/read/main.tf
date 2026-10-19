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
  skip_destroy    = false
}

data "elasticstack_fleet_agent_policy" "test" {
  name = elasticstack_fleet_agent_policy.test_policy.name
}
