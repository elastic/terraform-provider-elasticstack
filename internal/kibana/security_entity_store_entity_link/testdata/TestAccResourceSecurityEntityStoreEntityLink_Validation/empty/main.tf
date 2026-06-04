provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_entity_store_entity_link" "empty" {
  target_id  = "generic:acc-test-target"
  entity_ids = []
}
