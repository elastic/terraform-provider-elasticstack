data "elasticstack_kibana_security_entity_store_entities" "test" {
  entity_id = "generic:acc-test-entity"
  filter    = "entity.type:generic"
}
