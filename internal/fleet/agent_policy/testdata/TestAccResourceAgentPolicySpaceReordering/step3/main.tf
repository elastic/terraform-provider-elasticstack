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
  name             = var.policy_name
  namespace        = "default"
  description      = "Test space reordering - step 3: reorder spaces"
  monitor_logs     = true
  monitor_metrics  = false
  skip_destroy     = false
  # CRITICAL TEST: Reordering spaces (default now first)
  # With the fix: Still uses "default", resource found, updates in-place
  space_ids        = ["default", "space-test-a"]

  depends_on = [elasticstack_kibana_space.test_space]
}