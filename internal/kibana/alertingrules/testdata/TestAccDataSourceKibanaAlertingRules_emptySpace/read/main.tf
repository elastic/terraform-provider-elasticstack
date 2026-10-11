variable "space_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = var.space_id
}

data "elasticstack_kibana_alerting_rules" "test" {
  space_id = elasticstack_kibana_space.test.space_id
}
