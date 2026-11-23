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
  skip_destroy    = false
  global_data_tags = {
    tag1 = {
      string_value = "value1"
    }
    tag2 = {
      number_value = 1.1
    }
  }
}

data "elasticstack_fleet_agent_policy" "test" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
