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
  description     = "Test space reordering - step 4: revert space_ids to default"
  monitor_logs    = true
  monitor_metrics = false
  skip_destroy    = false
  # Revert to the default space by setting it explicitly. space_ids = [] is not
  # "unspecified": Fleet 9.1+ keeps the previous spaces and the provider then
  # adopts them, which produces an inconsistent apply.
  space_ids = ["default"]

  depends_on = [elasticstack_kibana_space.test_space]
}
