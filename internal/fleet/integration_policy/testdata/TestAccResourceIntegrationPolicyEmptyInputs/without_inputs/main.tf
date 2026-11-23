variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

variable "description" {
  description = "The integration policy description"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name      = var.policy_name
  namespace = "default"
}

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name                = "${var.policy_name}-integration"
  namespace           = "default"
  description         = var.description
  agent_policy_id     = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = "tcp"
  integration_version = "1.16.0"
}
