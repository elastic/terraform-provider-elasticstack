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
  entity_types = ["host"]
}

resource "elasticstack_kibana_security_entity_store_entity" "test" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  space_id    = elasticstack_kibana_space.test.space_id
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
