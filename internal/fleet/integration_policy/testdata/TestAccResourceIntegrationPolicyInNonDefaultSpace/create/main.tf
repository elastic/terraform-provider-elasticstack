variable "policy_name" {
  type = string
}

variable "space_id" {
  type = string
}

variable "space_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = var.space_id
  name        = var.space_name
  description = "Test space for Fleet integration policy space inheritance test"
}

resource "elasticstack_fleet_integration" "test" {
  name     = "tcp"
  version  = "1.16.0"
  force    = true
  space_id = var.space_id

  depends_on = [elasticstack_kibana_space.test_space]
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "${var.policy_name} Agent Policy"
  namespace       = "default"
  description     = "Test Agent Policy in Non-Default Space"
  monitor_logs    = true
  monitor_metrics = false
  skip_destroy    = false
  space_ids       = [var.space_id]

  depends_on = [elasticstack_kibana_space.test_space]
}

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name                = var.policy_name
  namespace           = "default"
  description         = "Test Integration Policy with inherited space"
  agent_policy_id     = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = elasticstack_fleet_integration.test.name
  integration_version = elasticstack_fleet_integration.test.version
  # space_ids is intentionally omitted to test automatic inheritance from the agent policy
}
