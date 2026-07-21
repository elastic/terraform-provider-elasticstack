provider "elasticstack" {
  kibana {}
}

# Fleet managed integration: Cloud Security Posture Management (CSPM) for an
# AWS account. Elastic provisions and runs the agent runtime in its own cloud
# infrastructure -- no Elastic Agent host is required.
#
# This resource is only supported on Elastic Cloud Hosted and Serverless
# (Security or Observability) deployments running Kibana 9.5.0+.
resource "elasticstack_fleet_managed_integration" "cspm_aws" {
  name            = "Agentless CSPM - AWS Production"
  description     = "Cloud Security Posture Management for the AWS production account"
  policy_template = "cspm"

  package = {
    name    = "cloud_security_posture"
    version = "3.4.0"
  }

  # Integration-level variables, as a JSON-encoded string.
  vars_json = jsonencode({
    posture    = "cspm"
    deployment = "aws"
  })

  # Selects which variable group ("deployment" flavor) applies at the
  # top level.
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

  # References an existing cloud connector (created and managed outside of
  # this resource, e.g. via the Fleet UI or API) for cross-account access
  # instead of static credentials. To actually route through the connector,
  # set the input's own "aws.credentials.type" var to "cloud_connectors" and
  # supply a matching "aws.credentials.external_id" (all sub-fields of
  # cloud_connector force replacement on change).
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
