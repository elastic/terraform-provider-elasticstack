variable "name" {
  type = string
}

variable "proxy_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_proxy" "test_proxy" {
  proxy_id = var.proxy_id
  name     = var.name
  url      = "https://proxy-explicit.example.com:3128"
}
