resource "elasticstack_kibana_security_entity_store" "test" {
  log_extraction = {
    delay     = "5m"
    frequency = "10m"
  }
}
