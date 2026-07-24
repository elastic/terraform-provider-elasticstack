variable "suffix" {
  type = string
}

variable "space_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-synthetics-param-${var.space_id}"
  description = "Kibana space for synthetics parameter acceptance test"
}

resource "elasticstack_kibana_synthetics_parameter" "test" {
  space_id    = elasticstack_kibana_space.test.space_id
  key         = "test-key-space-updated-${var.suffix}"
  value       = "test-value-space-updated"
  description = "Updated description in space"
  tags        = ["space-c"]
}
