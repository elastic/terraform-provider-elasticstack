provider "elasticstack" {
  kibana {}
}

variable "space_id" {
  type = string
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-entity-store-${var.space_id}"
  description = "Kibana space for entity store acceptance test"
}

resource "elasticstack_kibana_security_entity_store_entity_link" "self_link" {
  space_id   = elasticstack_kibana_space.test.space_id
  target_id  = "generic:acc-test-target"
  entity_ids = ["generic:acc-test-target", "generic:acc-test-alias"]
}
