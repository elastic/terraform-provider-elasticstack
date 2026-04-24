variable "kibana_endpoints" {
  description = "Kibana base URLs for the entity-local connection block"
  type        = list(string)
}

variable "api_key" {
  type    = string
  default = ""
}

variable "username" {
  type    = string
  default = ""
}

variable "password" {
  type    = string
  default = ""
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# install tcp integration via Fleet/Kibana: entity-local kibana_connection hits scoped client
resource "elasticstack_fleet_integration" "test_integration" {
  name         = "tcp"
  version      = "1.16.0"
  force        = true
  skip_destroy = true

  kibana_connection {
    endpoints = var.kibana_endpoints
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
  }
}
