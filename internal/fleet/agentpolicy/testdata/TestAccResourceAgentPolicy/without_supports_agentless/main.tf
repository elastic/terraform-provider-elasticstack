provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy without supports_agentless"
  monitor_logs    = false
  monitor_metrics = true
  skip_destroy    = var.skip_destroy
}
