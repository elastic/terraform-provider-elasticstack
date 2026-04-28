provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name    = "Test Artifacts Source"
  host    = "https://artifacts.elastic.co/downloads/"
  default = true
}

resource "elasticstack_fleet_agent_policy" "test" {
  name        = "Test Agent Policy for Enrollment Tokens (No PolicyID)"
  namespace   = "default"
  description = "Agent Policy for testing Enrollment Tokens without policy_id filter"

  depends_on = [elasticstack_fleet_agent_download_source.default]
}

# Read all enrollment tokens without filtering by policy_id.
# This exercises the code path that lists every enrollment token across all policies.
data "elasticstack_fleet_enrollment_tokens" "all" {
  depends_on = [elasticstack_fleet_agent_policy.test]
}
