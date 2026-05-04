variable "space_a" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "space_a" {
  space_id = var.space_a
  name     = var.space_a
}

resource "elasticstack_fleet_integration" "test_a" {
  name         = "tcp"
  version      = "1.16.0"
  force        = true
  skip_destroy = false
  space_id     = elasticstack_kibana_space.space_a.space_id
}
