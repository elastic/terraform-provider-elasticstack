provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_entity_store" "store" {
  entity_types = ["generic"]
}

resource "elasticstack_kibana_security_entity_store_entity" "target" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  entity_type = "generic"
  entity_id   = "generic:acc-test-target"

  entity = {
    id     = "generic:acc-test-target"
    name   = "acc-test-target"
    type   = "generic"
    source = ["terraform-acc-test"]
  }
}

data "elasticstack_kibana_security_entity_store_resolution_group" "test" {
  depends_on = [elasticstack_kibana_security_entity_store_entity.target]

  entity_id = "generic:acc-test-target"
  space_id  = "default"
}
