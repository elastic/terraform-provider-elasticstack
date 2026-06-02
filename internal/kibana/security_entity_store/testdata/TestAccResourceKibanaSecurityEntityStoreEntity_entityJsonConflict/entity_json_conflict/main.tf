terraform {
  required_providers {
    elasticstack = {
      source = "elastic/elasticstack"
    }
  }
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_entity_store_entity" "test" {
  entity_type = "host"
  entity_id   = "host:acc-test-host-conflict"

  entity {
    id   = "host:acc-test-host-conflict"
    name = "conflict-test"
    type = "host"
  }

  entity_json = jsonencode({
    id   = "host:acc-test-host-conflict"
    name = "conflict-test"
    type = "host"
  })
}