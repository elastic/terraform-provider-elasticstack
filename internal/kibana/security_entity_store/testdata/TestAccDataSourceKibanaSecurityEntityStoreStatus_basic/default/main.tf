resource "elasticstack_kibana_security_entity_store" "test" {}

data "elasticstack_kibana_security_entity_store_status" "test" {
  depends_on = [elasticstack_kibana_security_entity_store.test]
}
