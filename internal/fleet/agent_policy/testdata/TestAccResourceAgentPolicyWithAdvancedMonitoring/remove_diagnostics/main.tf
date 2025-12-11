provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy - Diagnostics Removed"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = var.skip_destroy

  advanced_monitoring_options = {
    http_monitoring_endpoint = {
      enabled        = true
      host           = "0.0.0.0"
      port           = 8080
      buffer_enabled = true
      pprof_enabled  = true
    }
  }
}

