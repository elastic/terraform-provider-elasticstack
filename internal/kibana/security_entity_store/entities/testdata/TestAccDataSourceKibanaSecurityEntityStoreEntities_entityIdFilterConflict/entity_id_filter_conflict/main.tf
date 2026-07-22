variable "space_id" {
  type = string
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-entity-store-${var.space_id}"
  description = "Kibana space for entity store acceptance test"
}

data "elasticstack_kibana_security_entity_store_entities" "test" {
  space_id  = elasticstack_kibana_space.test.space_id
  entity_id = "generic:acc-test-entity"
  filter    = "entity.type:generic"
}
