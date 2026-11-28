variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

variable "capture_personal_data" {
  description = "Value of the APM capture_personal_data input var"
  type        = bool
  default     = true
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# The registry may only list prerelease versions of the APM package for the
# stack under test, so fall back to the latest prerelease when no GA version
# is available.
data "elasticstack_fleet_integration" "apm" {
  name = "apm"
}

data "elasticstack_fleet_integration" "apm_prerelease" {
  name       = "apm"
  prerelease = true
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
  integration_version = coalesce(data.elasticstack_fleet_integration.apm.version, data.elasticstack_fleet_integration.apm_prerelease.version)

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
