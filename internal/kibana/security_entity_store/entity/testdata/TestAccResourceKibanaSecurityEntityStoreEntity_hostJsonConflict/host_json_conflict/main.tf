resource "elasticstack_kibana_security_entity_store" "store" {
  entity_types = ["host"]
}

resource "elasticstack_kibana_security_entity_store_entity" "test" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  entity_type = "host"
  entity_id   = "host:acc-test-host-conflict"

  entity = {
    id   = "host:acc-test-host-conflict"
    name = "acc-test-host-conflict"
    type = "host"
  }

  host = {
    name = "acc-test-host-conflict"
  }

  host_json = jsonencode({
    name = "acc-test-host-conflict"
  })
}
