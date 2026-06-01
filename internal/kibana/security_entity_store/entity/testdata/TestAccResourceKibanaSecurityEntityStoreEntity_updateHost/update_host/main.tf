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
  entity_id   = "host:acc-test-host-01"

  entity {
    id   = "host:acc-test-host-01"
    name = "acc-test-host-01"
    type = "host"
    source = "terraform-acc-test"
  }

  host {
    name = "acc-test-host-01"
    ip   = ["10.0.1.42", "10.0.1.43"]
  }
}