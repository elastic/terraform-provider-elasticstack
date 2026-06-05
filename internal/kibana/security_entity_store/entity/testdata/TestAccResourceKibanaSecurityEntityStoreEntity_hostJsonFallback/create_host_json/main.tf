resource "elasticstack_kibana_security_entity_store" "store" {
  entity_types = ["host"]
}

resource "elasticstack_kibana_security_entity_store_entity" "test" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  entity_type = "host"
  entity_id   = "host:acc-test-host-json"

  entity = {
    id   = "host:acc-test-host-json"
    name = "acc-test-host-json"
    type = "host"
  }

  host_json = jsonencode({
    name = "acc-test-host-json"
    ip   = ["10.0.0.1"]
  })
}
