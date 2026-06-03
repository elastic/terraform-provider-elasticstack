data "elasticstack_kibana_security_entity_store_entities" "test" {
  sort_field = "entity.id"
  sort_order = "asc"
  filter     = "entity.type:generic"
}
