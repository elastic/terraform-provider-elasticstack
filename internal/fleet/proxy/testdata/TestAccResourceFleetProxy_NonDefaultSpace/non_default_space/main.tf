variable "name" {
  type = string
}

variable "space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Fleet Proxy ${var.space_id}"
  description = "Kibana space for fleet proxy acceptance test"
}

resource "elasticstack_fleet_proxy" "test_proxy" {
  space_id = elasticstack_kibana_space.test.space_id
  name     = var.name
  url      = "https://proxy-space.example.com:3128"
}
