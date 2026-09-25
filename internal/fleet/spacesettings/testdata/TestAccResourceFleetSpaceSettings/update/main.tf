variable "space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = var.space_id
}

resource "elasticstack_fleet_space_settings" "test" {
  space_id                   = elasticstack_kibana_space.test.space_id
  allowed_namespace_prefixes = ["team_a", "shared"]
}
