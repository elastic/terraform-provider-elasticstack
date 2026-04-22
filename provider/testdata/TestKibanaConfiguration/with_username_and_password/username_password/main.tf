variable "kibana_endpoint" {
  type = string
}

variable "kibana_username" {
  type = string
}

variable "kibana_password" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {
    endpoints = [var.kibana_endpoint]
    username  = var.kibana_username
    password  = var.kibana_password
  }
}

resource "elasticstack_kibana_space" "acc_test" {
  space_id = "acc_test_space"
  name     = "Acceptance Test Space"
}
