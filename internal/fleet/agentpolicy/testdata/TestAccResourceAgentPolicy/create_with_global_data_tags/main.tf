provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy"
  monitor_logs    = true
  monitor_metrics = false
  skip_destroy    = var.skip_destroy
  global_data_tags = {
    tag1 = {
      string_value = "value1"
    }
    tag2 = {
      number_value = 1.1
    }
  }
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}