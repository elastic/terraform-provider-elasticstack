provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "Test Policy"
  namespace       = "default"
  description     = "Test Agent Policy"
  sys_monitoring  = true
  monitor_logs    = true
  monitor_metrics = true

  global_data_tags = {
    first_tag = {
      string_value = "tag_value"
    },
    second_tag = {
      number_value = 1.2
    }
  }
}
