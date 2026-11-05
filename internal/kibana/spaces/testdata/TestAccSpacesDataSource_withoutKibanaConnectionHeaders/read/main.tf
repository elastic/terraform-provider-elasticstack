variable "kibana_endpoints" {
  type = list(string)
}

variable "api_key" {
  type      = string
  sensitive = true
  default   = ""
}

variable "username" {
  type    = string
  default = ""
}

variable "password" {
  type      = string
  sensitive = true
  default   = ""
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_kibana_spaces" "all_spaces" {
  kibana_connection {
    endpoints = var.kibana_endpoints
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
  }
}
