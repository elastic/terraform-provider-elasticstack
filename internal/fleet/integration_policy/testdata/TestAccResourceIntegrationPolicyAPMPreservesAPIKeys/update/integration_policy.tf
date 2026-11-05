variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

variable "capture_personal_data" {
  description = "Value of the APM capture_personal_data input var"
  type        = bool
  default     = false
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_integration" "apm" {
  name = "apm"
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "${var.policy_name} Agent Policy"
  namespace       = "default"
  description     = "IntegrationPolicyTest APM Agent Policy"
  monitor_logs    = false
  monitor_metrics = false
  skip_destroy    = false
}

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name                = var.policy_name
  namespace           = "default"
  description         = "IntegrationPolicyTest APM Policy"
  agent_policy_id     = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = data.elasticstack_fleet_integration.apm.name
  integration_version = data.elasticstack_fleet_integration.apm.version

  inputs = {
    "apmserver-apm" = {
      enabled = true
      vars = jsonencode({
        "host" : "0.0.0.0:8200"
        "url" : "http://0.0.0.0:8200"
        "enable_rum" : true
        "capture_personal_data" : var.capture_personal_data
      })
    }
  }
}
