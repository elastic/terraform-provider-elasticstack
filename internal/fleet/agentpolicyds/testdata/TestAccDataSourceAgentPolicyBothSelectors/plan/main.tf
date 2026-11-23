provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_agent_policy" "test" {
  policy_id = "abc123"
  name      = "both-selectors"
}
