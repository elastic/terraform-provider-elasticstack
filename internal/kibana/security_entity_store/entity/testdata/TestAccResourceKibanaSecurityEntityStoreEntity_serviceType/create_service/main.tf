resource "elasticstack_kibana_security_entity_store" "store" {
  entity_types = ["service"]
}

resource "elasticstack_kibana_security_entity_store_entity" "test" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

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
