variable "space_id_1" {
  type = string
}

variable "space_id_2" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test1" {
  space_id = var.space_id_1
  name     = var.space_id_1
}

resource "elasticstack_kibana_space" "test2" {
  space_id = var.space_id_2
  name     = var.space_id_2
}

resource "elasticstack_fleet_space_settings" "test" {
  space_id                   = elasticstack_kibana_space.test1.space_id
  allowed_namespace_prefixes = ["team_a"]
}
