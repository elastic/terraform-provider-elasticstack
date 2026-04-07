variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "test_policy_default" {
  name            = "Private Location Agent Policy - test_policy_default - ${var.suffix}"
  namespace       = "default"
  description     = "TestPrivateLocationResource Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "test" {
  label           = "pl-test-label-2-${var.suffix}"
  agent_policy_id = elasticstack_fleet_agent_policy.test_policy_default.policy_id
  geo = {
    lat = -33.21
    lon = -33.21
  }
}
