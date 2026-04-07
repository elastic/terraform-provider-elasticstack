variable "kibana_endpoint" {
  type = string
}

variable "kibana_api_key" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {
    endpoints = [var.kibana_endpoint]
    api_key   = var.kibana_api_key
  }
}

resource "elasticstack_kibana_space" "acc_test" {
  space_id = "acc_test_space"
  name     = "Acceptance Test Space"
}
