variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

variable "secret_key" {
  description = "The secret key for access"
  type        = string
}

variable "integration_name" {
  description = "The integration name"
  type        = string
  default     = "aws_logs"
}

variable "integration_version" {
  description = "The integration version"
  type        = string
  default     = "1.4.0"
}

variable "default_region" {
  description = "AWS default region"
  type        = string
  default     = "us-east-1"
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

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name                = var.policy_name
  namespace           = "default"
  description         = "IntegrationPolicyTest Policy"
  agent_policy_id     = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = elasticstack_fleet_integration.test_policy.name
  integration_version = elasticstack_fleet_integration.test_policy.version

  vars_json = jsonencode({
    "access_key_id" : "placeholder"
    "secret_access_key" : "${var.secret_key} ${var.policy_name}"
    "session_token" : "placeholder"
    "endpoint" : "endpoint"
    "default_region" : var.default_region
  })

  input {
    input_id = "aws_logs-aws-cloudwatch"
    enabled  = true
    streams_json = jsonencode({
      "aws_logs.generic" = {
        enabled = true
        vars = {
          "number_of_workers" : 1
          "log_streams" : []
          "start_position" : "beginning"
          "scan_frequency" : "1m"
          "api_timeput" : "120s"
          "api_sleep" : "200ms"
          "tags" : ["forwarded"]
          "preserve_original_event" : false
          "data_stream.dataset" : "aws_logs.generic"
          "custom" : ""
        }
      }
    })
  }

  input {
    input_id = "aws_logs-aws-s3"
    enabled  = true
    streams_json = jsonencode({
      "aws_logs.generic" = {
        enabled = true
        vars = {
          "number_of_workers" : 1
          "bucket_list_interval" : "120s"
          "file_selectors" : ""
          "fips_enabled" : false
          "include_s3_metadata" : []
          "max_bytes" : "10MiB"
          "max_number_of_messages" : 5
          "parsers" : ""
          "sqs.max_receive_count" : 5
          "sqs.wait_time" : "20s"
          "tags" : ["forwarded"]
          "preserve_original_event" : false
          "data_stream.dataset" : "aws_logs.generic"
          "custom" : ""
        }
      }
    })
  }
}