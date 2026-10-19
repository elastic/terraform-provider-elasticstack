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
    internal = <<-YAML
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

data "elasticstack_fleet_agent_policy" "test" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
