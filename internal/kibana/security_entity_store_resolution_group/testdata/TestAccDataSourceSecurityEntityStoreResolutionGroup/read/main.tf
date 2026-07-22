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

resource "elasticstack_kibana_security_entity_store" "store" {
  space_id     = elasticstack_kibana_space.test.space_id
  entity_types = ["generic"]
}

resource "elasticstack_kibana_security_entity_store_entity" "target" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  space_id    = elasticstack_kibana_space.test.space_id
  entity_type = "generic"
  entity_id   = "generic:acc-test-target"

  entity = {
    id     = "generic:acc-test-target"
    name   = "acc-test-target"
    type   = "generic"
    source = ["terraform-acc-test"]
  }
}

data "elasticstack_kibana_security_entity_store_resolution_group" "test" {
  depends_on = [elasticstack_kibana_security_entity_store_entity.target]

  entity_id = "generic:acc-test-target"
  space_id  = elasticstack_kibana_space.test.space_id
}
