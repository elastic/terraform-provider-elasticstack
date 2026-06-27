provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "unit-test-policy"
  namespace       = "default"
  policy_id       = "my..policy"
  monitor_logs    = true
  monitor_metrics = false
}
