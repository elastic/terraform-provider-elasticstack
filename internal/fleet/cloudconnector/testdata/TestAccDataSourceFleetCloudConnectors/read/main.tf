variable "space_id" {
  type = string
}

variable "empty_space_id" {
  type = string
}

variable "space_name" {
  type = string
}

variable "empty_space_name" {
  type = string
}

variable "aws_name_1" {
  type = string
}

variable "aws_name_2" {
  type = string
}

variable "azure_name" {
  type = string
}

variable "role_arn" {
  type = string
}

variable "aws_external_id_1" {
  type      = string
  sensitive = true
}

variable "aws_external_id_2" {
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

variable "azure_cloud_connector_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "connectors" {
  space_id    = var.space_id
  name        = var.space_name
  description = "Space for cloud connector data source acceptance test"
}

resource "elasticstack_kibana_space" "empty" {
  space_id    = var.empty_space_id
  name        = var.empty_space_name
  description = "Empty space for cloud connector data source acceptance test"
}

resource "elasticstack_fleet_cloud_connector" "aws1" {
  space_id       = elasticstack_kibana_space.connectors.space_id
  name           = var.aws_name_1
  cloud_provider = "aws"
  aws = {
    role_arn    = var.role_arn
    external_id = var.aws_external_id_1
  }
}

resource "elasticstack_fleet_cloud_connector" "aws2" {
  space_id       = elasticstack_kibana_space.connectors.space_id
  name           = var.aws_name_2
  cloud_provider = "aws"
  aws = {
    role_arn    = var.role_arn
    external_id = var.aws_external_id_2
  }
}

resource "elasticstack_fleet_cloud_connector" "azure" {
  space_id       = elasticstack_kibana_space.connectors.space_id
  name           = var.azure_name
  cloud_provider = "azure"
  azure = {
    tenant_id          = var.azure_tenant_id
    client_id          = var.azure_client_id
    cloud_connector_id = var.azure_cloud_connector_id
  }
}

data "elasticstack_fleet_cloud_connectors" "all" {
  space_id = elasticstack_kibana_space.connectors.space_id
}

data "elasticstack_fleet_cloud_connectors" "aws_only" {
  space_id = elasticstack_kibana_space.connectors.space_id
  kuery    = "fleet-cloud-connector.attributes.cloudProvider:aws"
}

data "elasticstack_fleet_cloud_connectors" "empty_space" {
  space_id = elasticstack_kibana_space.empty.space_id
}
