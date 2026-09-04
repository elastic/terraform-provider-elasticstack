provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = var.space_name
  description = "Test space for Fleet enrollment tokens space_id (no policy_id) test"
}

resource "elasticstack_fleet_agent_policy" "test" {
  name        = "Test Agent Policy for Enrollment Tokens (SpaceID, no policy_id)"
  namespace   = "default"
  description = "Agent Policy for testing Enrollment Tokens with space_id and no policy_id"
  space_ids   = [var.space_id]

  depends_on = [elasticstack_kibana_space.test]
}

# Read all enrollment tokens in the custom space, without filtering by policy_id.
# This exercises fleet.GetEnrollmentTokens with a non-default space.
data "elasticstack_fleet_enrollment_tokens" "test" {
  space_id = var.space_id

  depends_on = [elasticstack_fleet_agent_policy.test]
}
