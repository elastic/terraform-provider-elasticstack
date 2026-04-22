variable "space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = "Test Space"
}

resource "elasticstack_fleet_integration" "test_integration_upgrade" {
  name         = "tcp"
  version      = "1.16.0"
  space_id     = elasticstack_kibana_space.test.space_id
  force        = true
  skip_destroy = true
}
