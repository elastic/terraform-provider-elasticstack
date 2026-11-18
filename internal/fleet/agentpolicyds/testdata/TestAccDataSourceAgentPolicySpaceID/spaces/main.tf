provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = var.space_name
  description = "Test space for Fleet agent policy data source space_id"
}

resource "elasticstack_fleet_agent_policy" "default_space" {
  name         = var.policy_name
  namespace    = "default"
  description  = "Default space agent policy"
  skip_destroy = false
}

resource "elasticstack_fleet_agent_policy" "other_space" {
  name         = var.policy_name
  namespace    = "default"
  description  = "Non-default space agent policy"
  skip_destroy = false
  space_ids    = [var.space_id]

  depends_on = [elasticstack_kibana_space.test]
}

data "elasticstack_fleet_agent_policy" "omitted" {
  name = var.policy_name

  depends_on = [
    elasticstack_fleet_agent_policy.default_space,
    elasticstack_fleet_agent_policy.other_space,
  ]
}

data "elasticstack_fleet_agent_policy" "explicit_default" {
  name     = var.policy_name
  space_id = "default"

  depends_on = [
    elasticstack_fleet_agent_policy.default_space,
    elasticstack_fleet_agent_policy.other_space,
  ]
}

data "elasticstack_fleet_agent_policy" "other" {
  name     = var.policy_name
  space_id = var.space_id

  depends_on = [
    elasticstack_fleet_agent_policy.default_space,
    elasticstack_fleet_agent_policy.other_space,
  ]
}
