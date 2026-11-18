variable "policy_name" {
  type = string
}

variable "package_version" {
  type = string
}

variable "package_title" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_managed_integration" "test" {
  name            = var.policy_name
  description     = "Managed integration CSPM Package-Title Test Policy"
  policy_template = "cspm"

  package = {
    name    = "cloud_security_posture"
    version = var.package_version
    title   = var.package_title
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
            role_arn               = "arn:aws:iam::123456789012:role/tf-acc-test-role"
            "aws.credentials.type" = "assume_role"
            "aws.account_type"     = "single-account"
          })
        }
      }
    }
  }
}
