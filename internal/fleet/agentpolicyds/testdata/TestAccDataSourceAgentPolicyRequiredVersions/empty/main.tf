provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name              = var.policy_name
  namespace         = "default"
  description       = "Test Agent Policy with Required Versions"
  monitor_logs      = true
  monitor_metrics   = false
  skip_destroy      = false
  required_versions = {}
}

data "elasticstack_fleet_agent_policy" "test" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
