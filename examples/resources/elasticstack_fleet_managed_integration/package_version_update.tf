provider "elasticstack" {
  kibana {}
}

# Bump package.version on an existing managed integration to pick up a newer Fleet
# package release. Terraform applies this as an in-place update (PUT to
# /api/fleet/managed_integrations) rather than destroy-and-recreate.
# package.name must stay the same; changing it still forces replacement.
resource "elasticstack_fleet_managed_integration" "cspm_package_upgrade" {
  name            = "CSPM package upgrade example"
  policy_template = "cspm"

  package = {
    name    = "cloud_security_posture"
    version = "3.5.0" # change from e.g. "3.4.0" to upgrade in place
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
}
