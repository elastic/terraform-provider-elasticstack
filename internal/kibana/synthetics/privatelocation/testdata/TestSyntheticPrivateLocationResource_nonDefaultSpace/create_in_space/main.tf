variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "Private Location Agent Policy - test_policy - ${var.suffix}"
  namespace       = "testacc"
  description     = "TestPrivateLocationResource Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "test" {
  space_id        = "testacc"
  label           = "pl-test-label-space-${var.suffix}"
  agent_policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
  tags            = ["a", "b"]
  geo = {
    lat = 42.42
    lon = -42.42
  }
}
