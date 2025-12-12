provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy with Default Advanced Settings"
  monitor_logs    = true
  monitor_metrics = true

  # Empty block - schema defaults are applied for flat attributes
  advanced_settings = {}
}

