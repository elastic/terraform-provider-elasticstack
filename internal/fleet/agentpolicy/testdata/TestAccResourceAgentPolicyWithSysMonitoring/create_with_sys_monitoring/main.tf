provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy with sys_monitoring"
  monitor_logs    = false
  monitor_metrics = false
  sys_monitoring  = true
}
