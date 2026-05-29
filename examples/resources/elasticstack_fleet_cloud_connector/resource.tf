variable "aws_external_id" {
  type      = string
  sensitive = true
}

variable "azure_tenant_id" {
  type      = string
  sensitive = true
}

variable "azure_client_id" {
  type      = string
  sensitive = true
}

variable "gcp_service_account_key" {
  type      = string
  sensitive = true
}

provider "elasticstack" {
  kibana {}
}

# Typed AWS block: compiles to the same wire vars payload and is repopulated in state after Read.
resource "elasticstack_fleet_cloud_connector" "aws_example" {
  name           = "Production AWS Connector"
  cloud_provider = "aws"
  account_type   = "single-account"

  aws = {
    role_arn    = "arn:aws:iam::123456789012:role/ElasticFleetConnector"
    external_id = var.aws_external_id
  }
}

# Typed Azure block: tenant_id and client_id are write-only secrets managed like aws.external_id.
resource "elasticstack_fleet_cloud_connector" "azure_example" {
  name           = "Production Azure Connector"
  cloud_provider = "azure"
  account_type   = "organization-account"

  azure = {
    tenant_id          = var.azure_tenant_id
    client_id          = var.azure_client_id
    cloud_connector_id = "azure-connector-prod-001"
  }
}

# Generic vars map for GCP (or custom integrations): use structured password vars for secrets.
resource "elasticstack_fleet_cloud_connector" "gcp_example" {
  name           = "Production GCP Connector"
  cloud_provider = "gcp"

  vars = {
    service_account = {
      type  = "text"
      value = "fleet-connector@my-project.iam.gserviceaccount.com"
    }
    audience = {
      type  = "text"
      value = "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/elastic/providers/fleet"
    }
    gcp_credentials_cloud_connector_id = {
      type  = "text"
      value = "gcp-connector-prod-001"
    }
    service_account_key = {
      type         = "password"
      secret_value = var.gcp_service_account_key
    }
  }
}
