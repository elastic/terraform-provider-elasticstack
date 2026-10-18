provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_agent_policy" "test" {
  name     = var.policy_name
  space_id = "default"
}
