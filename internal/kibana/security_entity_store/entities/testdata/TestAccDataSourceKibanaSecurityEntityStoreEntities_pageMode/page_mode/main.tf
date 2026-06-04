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

  page         = 1
  per_page     = 10
  sort_field   = "entity.id"
  sort_order   = "asc"
  entity_types = ["generic"]
}
