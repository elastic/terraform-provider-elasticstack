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

data "elasticstack_kibana_security_entity_store_resolution_group" "test" {
  depends_on = [elasticstack_kibana_security_entity_store_entity.target]

  entity_id = "user:target@example.com"
}
