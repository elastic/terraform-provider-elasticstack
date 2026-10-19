provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name         = var.policy_name
  namespace    = "default"
  description  = "Test Agent Policy"
  skip_destroy = false
}

data "elasticstack_fleet_agent_policy" "test" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
