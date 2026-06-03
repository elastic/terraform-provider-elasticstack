provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_entity_store" "store" {
  entity_types = ["user"]
}

resource "elasticstack_kibana_security_entity_store_entity" "target" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  entity_type = "user"
  entity_id   = "user:target@example.com"

  entity = {
    id     = "user:target@example.com"
    name   = "target"
    type   = "user"
    source = ["terraform-acc-test"]
  }
}

resource "elasticstack_kibana_security_entity_store_entity" "alias1" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  entity_type = "user"
  entity_id   = "user:alias1@example.com"

  entity = {
    id     = "user:alias1@example.com"
    name   = "alias1"
    type   = "user"
    source = ["terraform-acc-test"]
  }
}

resource "elasticstack_kibana_security_entity_store_entity" "alias2" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  entity_type = "user"
  entity_id   = "user:alias2@example.com"

  entity = {
    id     = "user:alias2@example.com"
    name   = "alias2"
    type   = "user"
    source = ["terraform-acc-test"]
  }
}

resource "elasticstack_kibana_security_entity_store_entity_link" "test" {
  depends_on = [
    elasticstack_kibana_security_entity_store_entity.target,
    elasticstack_kibana_security_entity_store_entity.alias1,
    elasticstack_kibana_security_entity_store_entity.alias2,
  ]

  target_id  = "user:target@example.com"
  entity_ids = ["user:alias1@example.com", "user:alias2@example.com"]
}
