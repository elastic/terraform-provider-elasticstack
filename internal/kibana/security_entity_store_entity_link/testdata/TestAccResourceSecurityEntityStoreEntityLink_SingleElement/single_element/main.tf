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

resource "elasticstack_kibana_security_entity_store_entity" "alias1" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  entity_type = "generic"
  entity_id   = "generic:acc-test-alias1"

  entity = {
    id     = "generic:acc-test-alias1"
    name   = "acc-test-alias1"
    type   = "generic"
    source = ["terraform-acc-test"]
  }
}

resource "elasticstack_kibana_security_entity_store_entity_link" "test" {
  depends_on = [
    elasticstack_kibana_security_entity_store_entity.target,
    elasticstack_kibana_security_entity_store_entity.alias1,
  ]

  target_id  = "generic:acc-test-target"
  entity_ids = toset(["generic:acc-test-alias1"])
}
