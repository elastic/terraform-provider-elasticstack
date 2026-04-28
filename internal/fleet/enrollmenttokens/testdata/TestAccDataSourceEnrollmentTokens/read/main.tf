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
  policy_id   = "223b1bf8-240f-463f-8466-5062670d0754"
  name        = "Test Agent Policy"
  namespace   = "default"
  description = "Agent Policy for testing Enrollment Tokens"

  depends_on = [elasticstack_fleet_agent_download_source.default]
}

data "elasticstack_fleet_enrollment_tokens" "test" {
  policy_id = elasticstack_fleet_agent_policy.test.policy_id
}
