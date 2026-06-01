variable "entity_id" {
  description = "The entity ID to look up"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

data "elasticstack_kibana_security_entity_store_resolution_group" "test" {
  entity_id = var.entity_id
}
