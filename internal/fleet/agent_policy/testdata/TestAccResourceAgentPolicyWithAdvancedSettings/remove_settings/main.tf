provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy without Advanced Settings"
  monitor_logs    = true
  monitor_metrics = true

  # advanced_settings removed entirely - UseStateForUnknown preserves state
}

