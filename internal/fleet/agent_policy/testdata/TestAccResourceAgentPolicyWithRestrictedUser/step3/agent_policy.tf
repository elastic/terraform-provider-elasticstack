provider "elasticstack" {
  alias = "restricted"
  elasticsearch {
    username = var.username
    password = var.password
  }
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  provider    = elasticstack.restricted
  name        = var.policy_name
  namespace   = "default"
  description = "Updated Test Agent Policy"
  space_ids   = [var.space_id]
}
