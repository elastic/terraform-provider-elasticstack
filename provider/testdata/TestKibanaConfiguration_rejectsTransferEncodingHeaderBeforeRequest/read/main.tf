variable "kibana_endpoint" {
  type = string
}

variable "header_name" {
  type = string
}

variable "header_value" {
  type      = string
  sensitive = true
}

provider "elasticstack" {
  kibana {
    endpoints = [var.kibana_endpoint]
    headers = {
      (var.header_name) = var.header_value
    }
  }
}

data "elasticstack_kibana_spaces" "all_spaces" {}
