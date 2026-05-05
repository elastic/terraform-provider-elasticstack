provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name               = var.policy_name
  namespace          = "default"
  description        = "Test Agent Policy with supports_agentless false"
  monitor_logs       = false
  monitor_metrics    = true
  skip_destroy       = var.skip_destroy
  supports_agentless = false
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
