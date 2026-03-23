variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

variable "output_name" {
  description = "The output name"
  type        = string
}

variable "updated_output_name" {
  description = "The updated output name"
  type        = string
}

variable "integration_name" {
  description = "The integration name"
  type        = string
  default     = "tcp"
}

variable "integration_version" {
  description = "The integration version"
  type        = string
  default     = "1.16.0"
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_policy" {
  name    = var.integration_name
  version = var.integration_version
  force   = true
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "${var.policy_name} Agent Policy"
  namespace       = "default"
  description     = "IntegrationPolicyTest Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

resource "elasticstack_fleet_output" "test_output" {
  name      = var.output_name
  output_id = "${var.policy_name}-test-output"
  type      = "elasticsearch"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "https://elasticsearch:9200"
  ]
}

resource "elasticstack_fleet_output" "test_output_updated" {
  name      = var.updated_output_name
  output_id = "${var.policy_name}-updated-output"
  type      = "elasticsearch"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "https://elasticsearch:9200"
  ]
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name                = var.policy_name
  namespace           = "default"
  description         = "Updated Integration Policy with Output"
  agent_policy_id     = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = elasticstack_fleet_integration.test_policy.name
  integration_version = elasticstack_fleet_integration.test_policy.version
  output_id           = elasticstack_fleet_output.test_output_updated.output_id

  inputs = {
    "tcp-tcp" = {
      enabled = true
      streams = {
        "tcp.generic" = {
          enabled = true
          vars = jsonencode({
            "listen_address" : "localhost"
            "listen_port" : 8080
            "data_stream.dataset" : "tcp.generic"
            "tags" : []
            "syslog_options" : "field: message"
            "ssl" : ""
            "custom" : ""
          })
        }
      }
    }
  }
}