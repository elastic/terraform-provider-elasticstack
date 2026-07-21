variable "policy_name" {
  type = string
}

variable "package_version" {
  type = string
}

variable "input_condition" {
  type = string
}

variable "stream_condition" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_managed_integration" "test" {
  name            = var.policy_name
  description     = "condition round-trip acceptance test"
  policy_template = "cspm"

  package = {
    name    = "cloud_security_posture"
    version = var.package_version
  }

  vars_json = jsonencode({
    posture    = "cspm"
    deployment = "aws"
  })

  var_group_selections = {
    deployment = "aws"
  }

  inputs = {
    "cspm-cloudbeat/cis_aws" = {
      enabled   = true
      condition = var.input_condition
      streams = {
        "cloud_security_posture.findings" = {
          enabled   = true
          condition = var.stream_condition
          vars = jsonencode({
            role_arn               = "arn:aws:iam::123456789012:role/tf-acc-test-role"
            "aws.credentials.type" = "assume_role"
            "aws.account_type"     = "single-account"
          })
        }
      }
    }
  }
}
