provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy with Default Advanced Monitoring"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = var.skip_destroy
  space_ids       = ["default"]

  # Empty nested blocks - schema defaults are applied for leaf attributes
  advanced_monitoring_options = {
    http_monitoring_endpoint = {}
    diagnostics = {
      rate_limits   = {}
      file_uploader = {}
    }
  }
}

