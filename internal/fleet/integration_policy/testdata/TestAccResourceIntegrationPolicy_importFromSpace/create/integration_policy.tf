variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

variable "space_id" {
  description = "The Kibana space ID to create the policy in"
  type        = string
}

variable "space_name" {
  description = "The Kibana space display name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = var.space_id
  name        = var.space_name
  description = "Test space for integration policy space import test"
}

resource "elasticstack_fleet_integration" "test_policy" {
  name    = "system"
  version = "1.64.0"
  force   = true
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "${var.policy_name} Agent Policy"
  namespace       = "default"
  description     = "IntegrationPolicyTest Agent Policy in Space"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
  space_ids       = [var.space_id]

  depends_on = [elasticstack_kibana_space.test_space]
}

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name                = var.policy_name
  namespace           = "default"
  description         = "Integration Policy in Space"
  agent_policy_id     = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = elasticstack_fleet_integration.test_policy.name
  integration_version = elasticstack_fleet_integration.test_policy.version
  space_ids           = [var.space_id]
}
