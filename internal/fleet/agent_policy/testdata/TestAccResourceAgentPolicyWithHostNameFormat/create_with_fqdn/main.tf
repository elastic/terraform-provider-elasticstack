provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name             = var.policy_name
  namespace        = "default"
  description      = "Test Agent Policy with FQDN host name format"
  monitor_logs     = true
  monitor_metrics  = false
  host_name_format = "fqdn"
}

