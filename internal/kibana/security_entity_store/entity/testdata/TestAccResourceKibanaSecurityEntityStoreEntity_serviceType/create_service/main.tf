variable "space_id" {
  type = string
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-entity-store-${var.space_id}"
  description = "Kibana space for entity store acceptance test"
}

resource "elasticstack_kibana_security_entity_store" "store" {
  space_id     = elasticstack_kibana_space.test.space_id
  entity_types = ["service"]
}

resource "elasticstack_kibana_security_entity_store_entity" "test" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  space_id    = elasticstack_kibana_space.test.space_id
  entity_type = "service"
  entity_id   = "service:acc-test-service"

  entity = {
    id   = "service:acc-test-service"
    name = "acc-test-service"
    type = "service"
  }

  service = {
    name = "acc-test-service"
  }
}
