variable "policy_name" {
  type = string
}

variable "package_version" {
  type = string
}

variable "space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Agentless Policy ${var.space_id}"
  description = "Kibana space for fleet agentless policy acceptance test"
}

resource "elasticstack_fleet_managed_integration" "test" {
  name            = var.policy_name
  description     = "Agentless CSPM Non-Default-Space Test Policy"
  policy_template = "cspm"
  space_ids       = [elasticstack_kibana_space.test.space_id]

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
            role_arn               = "arn:aws:iam::123456789012:role/tf-acc-test-role"
            "aws.credentials.type" = "assume_role"
            "aws.account_type"     = "single-account"
          })
        }
      }
    }
  }
}
