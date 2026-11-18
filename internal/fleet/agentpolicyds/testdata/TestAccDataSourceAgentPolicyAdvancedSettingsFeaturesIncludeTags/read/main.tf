provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy with features_include_tags_in_events_enabled"
  monitor_logs    = true
  monitor_metrics = true

  advanced_settings = {
    features_include_tags_in_events_enabled = true
  }
}

data "elasticstack_fleet_agent_policy" "test" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
