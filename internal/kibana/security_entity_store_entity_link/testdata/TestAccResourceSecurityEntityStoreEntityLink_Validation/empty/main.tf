provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_entity_store_entity_link" "empty" {
  target_id  = "user:target@example.com"
  entity_ids = []
}
