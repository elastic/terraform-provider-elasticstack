variable "policy_name" {
  type = string
}

variable "package_version" {
  type = string
}

variable "cloud_connector_id" {
  type = string
}

# external_id_secret_id is the ID of a Fleet secret already minted by the
# test (see mintExternalIDSecretRef in acc_test.go) that backs the cloud
# connector's own external_id. Configuring aws.credentials.external_id as
# the already-secret-ref-shaped value below (rather than a plaintext string)
# is a test-fixture-only workaround for a real gap: this resource does not
# yet implement secret-masking reconciliation for password-type vars, so a
# plaintext value here would cause "Provider produced inconsistent result
# after apply" once Kibana echoes it back as a {id,isSecretRef} object. See
# acc_test.go's TestAccResourceAgentlessPolicy_CloudConnector doc comment.
variable "external_id_secret_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_managed_integration" "test" {
  name            = var.policy_name
  description     = "Agentless CSPM Cloud Connector Test Policy"
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
            role_arn               = "arn:aws:iam::123456789012:role/tf-acc-test-role"
            "aws.credentials.type" = "cloud_connectors"
            "aws.account_type"     = "single-account"
            "aws.credentials.external_id" = {
              isSecretRef = true
              id          = var.external_id_secret_id
            }
          })
        }
      }
    }
  }
}
