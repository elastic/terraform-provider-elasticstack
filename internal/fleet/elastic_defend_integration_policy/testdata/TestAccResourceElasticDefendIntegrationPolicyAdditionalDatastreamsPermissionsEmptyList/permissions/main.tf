variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

variable "additional_datastreams_permissions" {
  description = "Additional data stream permissions; null leaves the attribute unset"
  type        = list(string)
  default     = null
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test" {
  name      = "${var.policy_name}-agent-policy"
  namespace = "default"
}

resource "elasticstack_fleet_elastic_defend_integration_policy" "test" {
  name                               = var.policy_name
  namespace                          = "default"
  agent_policy_id                    = elasticstack_fleet_agent_policy.test.policy_id
  enabled                            = true
  integration_version                = "8.14.0"
  preset                             = "NGAV"
  additional_datastreams_permissions = var.additional_datastreams_permissions

  policy = {
    windows = { malware = { mode = "prevent" } }
    mac     = { malware = { mode = "prevent" } }
    linux   = { malware = { mode = "detect" } }
  }
}
