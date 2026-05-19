variable "space_id" {
  description = "The space ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = var.space_id
  name        = format("SDK %s", var.space_id)
  description = "Space created by the legacy SDK provider"
}
