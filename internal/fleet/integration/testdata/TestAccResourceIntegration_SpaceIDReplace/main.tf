variable "space_id" {
  description = "The Kibana space ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = var.space_id
}

resource "elasticstack_fleet_integration" "test_integration" {
  name         = "tcp"
  version      = "1.16.0"
  force        = true
  skip_destroy = false
  space_id     = elasticstack_kibana_space.test.space_id
}
