resource "elasticstack_kibana_security_entity_store" "test" {
  history_snapshot = {
    frequency = "1d"
  }
}
