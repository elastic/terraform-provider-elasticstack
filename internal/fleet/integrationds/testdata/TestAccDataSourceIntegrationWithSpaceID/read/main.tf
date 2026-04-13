provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

variable "space_name" { type = string }
variable "space_id" { type = string }

resource "elasticstack_kibana_space" "test" {
  name     = var.space_name
  space_id = var.space_id
}

data "elasticstack_fleet_integration" "test" {
  name     = "tcp"
  space_id = elasticstack_kibana_space.test.space_id
}
