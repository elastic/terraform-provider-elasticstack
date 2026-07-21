variable "policy_name" {
  description = "The agentless policy name"
  type        = string
}

variable "package_version" {
  description = "The cloud_security_posture package version"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_managed_integration" "test" {
  name            = var.policy_name
  description     = "Updated Agentless CSPM Test Policy"
  policy_template = "cspm"

  package = {
    name    = "cloud_security_posture"
    version = var.package_version
  }

  vars_json = jsonencode({
    posture    = "cspm"
    deployment = "aws"
  })

  inputs = {
    "cspm-cloudbeat/cis_aws" = {
      enabled = true
      streams = {
        "cloud_security_posture.findings" = {
          enabled = true
          vars = jsonencode({
            role_arn               = "arn:aws:iam::123456789012:role/tf-acc-test-role-updated"
            "aws.credentials.type" = "assume_role"
            "aws.account_type"     = "organization-account"
          })
        }
      }
    }
  }
}
