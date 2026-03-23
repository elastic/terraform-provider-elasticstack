provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "This policy was not created due to bad tags"
  monitor_logs    = false
  monitor_metrics = true
  skip_destroy    = var.skip_destroy
  global_data_tags = {
    tag1 = {
      string_value = "value1a"
      number_value = 1.2
    }
  }
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}