resource "elasticstack_kibana_security_entity_store" "store" {
  entity_types = ["generic"]
}

resource "elasticstack_kibana_security_entity_store_entity" "test" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  entity_type = "generic"
  entity_id   = "generic:acc-test-entity"

  entity_json = jsonencode({
    id     = "generic:wrong-id"
    name   = "acc-test-entity"
    type   = "generic"
    source = ["terraform-acc-test"]
  })
}
