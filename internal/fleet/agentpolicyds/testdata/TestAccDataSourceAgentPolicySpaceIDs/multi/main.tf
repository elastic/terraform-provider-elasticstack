provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = var.space_name
  description = "Test space for Fleet agent policy data source space_ids"
}

resource "elasticstack_kibana_space" "second" {
  space_id    = var.second_space_id
  name        = var.second_space_name
  description = "Second test space for Fleet agent policy data source space_ids"
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy with multiple Space IDs"
  monitor_logs    = true
  monitor_metrics = false
  skip_destroy    = false
  space_ids       = [var.space_id, var.second_space_id]

  depends_on = [elasticstack_kibana_space.test, elasticstack_kibana_space.second]
}

data "elasticstack_fleet_agent_policy" "test" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
  space_id  = var.space_id
}
