variable "kibana_endpoint" {
  type = string
}

provider "elasticstack" {}

data "elasticstack_kibana_spaces" "test" {
  kibana_connection {
    endpoints = [var.kibana_endpoint]
    headers = {
      "X-Proxy-Token" = "first"
      "x-proxy-token" = "second"
    }
  }
}
