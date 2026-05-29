variable "name" {
  type = string
}

variable "tenant_id" {
  type      = string
  sensitive = true
}

variable "client_id" {
  type      = string
  sensitive = true
}

variable "cloud_connector_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_cloud_connector" "test" {
  name           = var.name
  cloud_provider = "azure"
  azure = {
    tenant_id          = var.tenant_id
    client_id          = var.client_id
    cloud_connector_id = var.cloud_connector_id
  }
}
