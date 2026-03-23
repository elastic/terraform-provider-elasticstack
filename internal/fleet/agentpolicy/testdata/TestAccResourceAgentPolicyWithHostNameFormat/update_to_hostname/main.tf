provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name             = var.policy_name
  namespace        = "default"
  description      = "Test Agent Policy with hostname format"
  monitor_logs     = true
  monitor_metrics  = false
  host_name_format = "hostname"
}

