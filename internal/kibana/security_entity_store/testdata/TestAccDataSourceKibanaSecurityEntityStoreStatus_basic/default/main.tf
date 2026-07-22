variable "space_id" {
  type = string
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-entity-store-${var.space_id}"
  description = "Kibana space for entity store acceptance test"
}

resource "elasticstack_kibana_security_entity_store" "test" {
  space_id = elasticstack_kibana_space.test.space_id
}

data "elasticstack_kibana_security_entity_store_status" "test" {
  space_id   = elasticstack_kibana_space.test.space_id
  depends_on = [elasticstack_kibana_security_entity_store.test]
}
