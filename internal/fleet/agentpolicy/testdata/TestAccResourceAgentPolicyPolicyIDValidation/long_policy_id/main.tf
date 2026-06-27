provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

locals {
  long_policy_id = join("", [for _ in range(256) : "a"])
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "unit-test-policy"
  namespace       = "default"
  policy_id       = local.long_policy_id
  monitor_logs    = true
  monitor_metrics = false
}
