provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy with Advanced Settings"
  monitor_logs    = true
  monitor_metrics = true

  advanced_settings = {
    logging_level    = "debug"
    logging_to_files = true
    go_max_procs     = 2
  }
}

