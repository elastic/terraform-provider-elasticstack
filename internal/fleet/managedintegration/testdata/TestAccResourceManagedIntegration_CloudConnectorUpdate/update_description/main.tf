variable "policy_name" {
  type = string
}

variable "package_version" {
  type = string
}

variable "cloud_connector_id" {
  type = string
}

variable "external_id_plaintext" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_managed_integration" "test" {
  name            = var.policy_name
  description     = "Updated description after cloud_connector association"
  policy_template = "cspm"

  package = {
    name    = "cloud_security_posture"
    version = var.package_version
  }

  vars_json = jsonencode({
    posture    = "cspm"
    deployment = "aws"
  })

  cloud_connector = {
    enabled            = true
    cloud_connector_id = var.cloud_connector_id
    target_csp         = "aws"
  }

  inputs = {
    "cspm-cloudbeat/cis_aws" = {
      enabled = true
      streams = {
        "cloud_security_posture.findings" = {
          enabled = true
          vars = jsonencode({
            role_arn                      = "arn:aws:iam::123456789012:role/tf-acc-test-role"
            "aws.credentials.type"        = "cloud_connectors"
            "aws.account_type"            = "single-account"
            "aws.credentials.external_id" = var.external_id_plaintext
          })
        }
      }
    }
  }
}
