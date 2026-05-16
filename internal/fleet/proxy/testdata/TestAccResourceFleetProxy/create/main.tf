variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_proxy" "test_proxy" {
  name = var.name
  url  = "https://proxy.example.com:3128"
}
