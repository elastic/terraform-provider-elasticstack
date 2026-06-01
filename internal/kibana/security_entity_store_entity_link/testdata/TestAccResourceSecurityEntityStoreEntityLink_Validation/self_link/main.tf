provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_entity_store_entity_link" "self_link" {
  target_id  = "user:target@example.com"
  entity_ids = ["user:target@example.com", "user:alias@example.com"]
}
