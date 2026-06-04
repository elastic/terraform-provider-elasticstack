resource "elasticstack_kibana_security_entity_store" "store" {
  entity_types = ["generic"]
}

resource "elasticstack_kibana_security_entity_store_entity" "test" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  entity_type = "generic"
  entity_id   = "generic:acc-test-entity"

  entity = {
    id     = "generic:acc-test-entity"
    name   = "acc-test-entity"
    type   = "generic"
    source = ["terraform-acc-test"]
  }
}

data "elasticstack_kibana_security_entity_store_entities" "test" {
  depends_on = [elasticstack_kibana_security_entity_store_entity.test]

  filter = "entity.id: \"generic:acc-test-entity\""
}
