resource "elasticstack_kibana_security_entity_store" "store" {
  entity_types = ["host"]
}

resource "elasticstack_kibana_security_entity_store_entity" "test" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  entity_type = "host"
  entity_id   = "host:acc-test-host"

  entity = {
    id   = "host:acc-test-host"
    name = "acc-test-host"
    type = "host"
  }

  host = {
    name = "acc-test-host"
    ip   = ["1.2.3.4"]
  }
}
