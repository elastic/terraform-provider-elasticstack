variable "policy_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test" {
  policy_id   = var.policy_id
  name        = "Test Agent Policy"
  namespace   = "default"
  description = "Agent Policy for testing Enrollment Tokens"
}

data "elasticstack_fleet_enrollment_tokens" "test" {
  policy_id = elasticstack_fleet_agent_policy.test.policy_id
}
