provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_entity_store_entity_link" "self_link" {
  target_id  = "generic:acc-test-target"
  entity_ids = ["generic:acc-test-target", "generic:acc-test-alias"]
}
