provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name                 = var.policy_name
  namespace            = "default"
  description          = "Test Agent Policy with Both Timeouts"
  monitor_logs         = false
  monitor_metrics      = true
  skip_destroy         = false
  inactivity_timeout   = "120s"
  unenrollment_timeout = "900s"
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}