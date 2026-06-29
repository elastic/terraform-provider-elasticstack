provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  policy_id       = var.policy_id
  description     = "Test Agent Policy with explicit policy_id"
  monitor_logs    = true
  monitor_metrics = false
  skip_destroy    = var.skip_destroy
}
