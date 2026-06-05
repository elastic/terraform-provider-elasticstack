resource "elasticstack_kibana_security_entity_store" "store" {
  entity_types = ["user"]
}

resource "elasticstack_kibana_security_entity_store_entity" "test" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  entity_type = "user"
  entity_id   = "user:acc-test-user@unknown"

  entity = {
    id   = "user:acc-test-user@unknown"
    name = "acc-test-user"
    type = "user"
  }

  user = {
    name = "acc-test-user"
  }
}
