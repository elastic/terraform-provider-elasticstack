provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = var.space_name
  description = "Test space for Fleet enrollment tokens space_id test"
}

resource "elasticstack_fleet_agent_policy" "test" {
  name        = "Test Agent Policy for Enrollment Tokens (SpaceID)"
  namespace   = "default"
  description = "Agent Policy for testing Enrollment Tokens with space_id"
  space_ids   = [var.space_id]

  depends_on = [elasticstack_kibana_space.test]
}

data "elasticstack_fleet_enrollment_tokens" "test" {
  policy_id = elasticstack_fleet_agent_policy.test.policy_id
  space_id  = var.space_id
}
