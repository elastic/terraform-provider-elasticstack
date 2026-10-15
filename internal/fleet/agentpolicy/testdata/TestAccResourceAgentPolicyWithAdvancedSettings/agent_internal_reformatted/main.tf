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
    internal = <<-YAML
      # reformatted but semantically equal
      agent:
          logging:
              level: info
          monitoring:
              http:
                  enabled: true
    YAML
  }
}
