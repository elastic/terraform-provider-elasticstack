resource "elasticstack_kibana_security_entity_store" "test" {
  entity_types             = ["host"]
  allow_entity_type_shrink = true
}
