provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy without explicit policy_id"
  monitor_logs    = true
  monitor_metrics = false
  skip_destroy    = var.skip_destroy
}
