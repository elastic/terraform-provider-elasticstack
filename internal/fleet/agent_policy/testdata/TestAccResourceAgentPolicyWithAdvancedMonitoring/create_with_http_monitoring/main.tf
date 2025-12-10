provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy with Advanced Monitoring"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = var.skip_destroy

  advanced_monitoring_options = {
    http_monitoring_endpoint = {
      enabled        = true
      host           = "localhost"
      port           = 6791
      buffer_enabled = false
      pprof_enabled  = false
    }
  }
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

