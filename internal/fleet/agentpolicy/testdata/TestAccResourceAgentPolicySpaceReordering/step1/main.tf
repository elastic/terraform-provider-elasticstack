provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = "space-test-a"
  name        = "Test Space A"
  description = "Test space for Fleet agent policy space reordering test"
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test space reordering - step 1"
  monitor_logs    = true
  monitor_metrics = false
  skip_destroy    = false
  space_ids       = ["default"]

  depends_on = [elasticstack_kibana_space.test_space]
}