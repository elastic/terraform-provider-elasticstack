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

resource "elasticstack_kibana_security_entity_store_entity" "alias1" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  space_id    = elasticstack_kibana_space.test.space_id
  entity_type = "generic"
  entity_id   = "generic:acc-test-alias1"

  entity = {
    id     = "generic:acc-test-alias1"
    name   = "acc-test-alias1"
    type   = "generic"
    source = ["terraform-acc-test"]
  }
}

resource "elasticstack_kibana_security_entity_store_entity_link" "test" {
  depends_on = [
    elasticstack_kibana_security_entity_store_entity.target,
    elasticstack_kibana_security_entity_store_entity.alias1,
  ]

  space_id   = elasticstack_kibana_space.test.space_id
  target_id  = "generic:acc-test-target"
  entity_ids = toset(["generic:acc-test-alias1"])
}
