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
  description     = "Test space reordering - step 2: prepend new space"
  monitor_logs    = true
  monitor_metrics = false
  skip_destroy    = false
  # CRITICAL TEST: Prepending "space-test-a" before "default"
  # Without the fix: Terraform queries using space-test-a, gets 404, recreates resource
  # With the fix: Terraform uses "default" (position-independent), finds resource, updates in-place
  space_ids = ["space-test-a", "default"]

  depends_on = [elasticstack_kibana_space.test_space]
}