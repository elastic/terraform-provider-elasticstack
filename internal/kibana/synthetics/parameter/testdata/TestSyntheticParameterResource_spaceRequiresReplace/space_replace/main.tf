variable "space_id" {
  type = string
}

variable "suffix" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-synthetics-param-replace-${var.space_id}"
  description = "Kibana space for synthetics parameter space_id replace test"
}

resource "elasticstack_kibana_synthetics_parameter" "test" {
  space_id = elasticstack_kibana_space.test.space_id
  key      = "test-key-space-replace-${var.suffix}"
  value    = "replace-target-value"
}
