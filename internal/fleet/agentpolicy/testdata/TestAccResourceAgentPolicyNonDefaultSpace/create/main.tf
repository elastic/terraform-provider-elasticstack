provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = var.space_id
  name        = var.space_name
  description = "Test space for Fleet agent policy non-default space test"
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy in Non-Default Space"
  monitor_logs    = true
  monitor_metrics = false
  skip_destroy    = false
  space_ids       = [var.space_id]

  depends_on = [elasticstack_kibana_space.test_space]
}
