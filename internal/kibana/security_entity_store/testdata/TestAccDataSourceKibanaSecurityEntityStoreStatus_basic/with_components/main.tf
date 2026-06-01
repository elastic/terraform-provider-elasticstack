resource "elasticstack_kibana_security_entity_store" "test" {}

data "elasticstack_kibana_security_entity_store_status" "test" {
  include_components = true
}
