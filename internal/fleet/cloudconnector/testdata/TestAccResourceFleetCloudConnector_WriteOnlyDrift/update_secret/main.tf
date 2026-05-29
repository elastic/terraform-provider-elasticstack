variable "name" {
  type = string
}

variable "role_arn" {
  type = string
}

variable "external_id" {
  type      = string
  sensitive = true
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_cloud_connector" "test" {
  name           = var.name
  cloud_provider = "aws"
  aws = {
    role_arn    = var.role_arn
    external_id = var.external_id
  }
}
