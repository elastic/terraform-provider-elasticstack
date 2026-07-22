variable "space_id" {
  type = string
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-entity-store-${var.space_id}"
  description = "Kibana space for entity store acceptance test"
}

resource "elasticstack_kibana_security_entity_store" "store" {
  space_id     = elasticstack_kibana_space.test.space_id
  entity_types = ["generic"]
}

resource "elasticstack_kibana_security_entity_store_entity" "test" {
  depends_on = [elasticstack_kibana_security_entity_store.store]

  space_id    = elasticstack_kibana_space.test.space_id
  entity_type = "generic"
  entity_id   = "generic:acc-test-entity"

  entity = {
    id     = "generic:acc-test-entity"
    name   = "acc-test-entity"
    type   = "generic"
    source = ["terraform-acc-test"]
  }
}

data "elasticstack_kibana_security_entity_store_entities" "test" {
  depends_on = [elasticstack_kibana_security_entity_store_entity.test]

  space_id     = elasticstack_kibana_space.test.space_id
  filter_query = "{\"match_all\":{}}"
  per_page     = 5
  page         = 1
  entity_types = ["generic"]
}
