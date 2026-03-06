provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name                 = var.policy_name
  namespace            = "default"
  description          = "Test Agent Policy with Unenrollment Timeout"
  monitor_logs         = true
  monitor_metrics      = false
  skip_destroy         = false
  unenrollment_timeout = "300s"
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}