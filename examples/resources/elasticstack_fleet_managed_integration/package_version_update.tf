provider "elasticstack" {
  kibana {}
}

# Standalone snippet: the same elasticstack_fleet_managed_integration.cspm_aws resource
# as in resource.tf after changing package.version in place (3.4.0 → 3.5.0). Apply this
# diff to an existing integration; do not add this file alongside resource.tf in one module.
resource "elasticstack_fleet_managed_integration" "cspm_aws" {
  name            = "Agentless CSPM - AWS Production"
  description     = "Cloud Security Posture Management for the AWS production account"
  policy_template = "cspm"

  package = {
    name    = "cloud_security_posture"
    version = "3.5.0" # was "3.4.0" — in-place update, not replacement
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
      enabled = true
      streams = {
        "cloud_security_posture.findings" = {
          enabled = true
          vars = jsonencode({
            role_arn               = "arn:aws:iam::123456789012:role/elastic-cspm-role"
            "aws.credentials.type" = "assume_role"
            "aws.account_type"     = "single-account"
          })
        }
      }
    }
  }

  cloud_connector = {
    enabled            = true
    cloud_connector_id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    name               = "aws-production-cross-account"
    target_csp         = "aws"
  }

  global_data_tags = {
    env = {
      string_value = "production"
    }
    team = {
      string_value = "cloud-security"
    }
  }

  additional_datastreams_permissions = ["logs-custom-*"]
}
