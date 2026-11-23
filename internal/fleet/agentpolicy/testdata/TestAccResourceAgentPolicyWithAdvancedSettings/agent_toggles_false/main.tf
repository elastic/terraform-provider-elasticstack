provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy with agent feature settings"
  monitor_logs    = true
  monitor_metrics = true

  advanced_settings = {
    features_disable_policy_change_acks_enabled = false
    features_include_tags_in_events_enabled     = false
  }
}
