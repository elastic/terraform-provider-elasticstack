variable "service_name" {
  description = "The APM service name"
  type        = string
}

variable "service_version" {
  description = "The APM service version"
  type        = string
}

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
  kibana {}
}

resource "elasticstack_apm_source_map" "test" {
  bundle_filepath = "/static/js/test.min.js"
  service_name    = var.service_name
  service_version = var.service_version
  sourcemap = {
    json = "{\"version\":3,\"file\":\"test.min.js\",\"sources\":[\"test.js\"],\"mappings\":\"AAAA\"}"
  }

  kibana_connection {
    endpoints = var.kibana_endpoints
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.username != "" ? var.username : null
    password  = var.password != "" ? var.password : null
  }
}
