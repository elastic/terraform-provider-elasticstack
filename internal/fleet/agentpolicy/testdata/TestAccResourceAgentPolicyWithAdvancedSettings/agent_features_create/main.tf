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
    features_disable_policy_change_acks_enabled = true
    features_include_tags_in_events_enabled     = true
    internal                                    = <<-YAML
      agent:
        monitoring:
          http:
            enabled: true
          # trailing comment
        logging:
          level: info
    YAML
  }
}
