provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy - No Advanced Monitoring"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = var.skip_destroy

  # advanced_monitoring_options removed entirely - UseStateForUnknown preserves state
}

