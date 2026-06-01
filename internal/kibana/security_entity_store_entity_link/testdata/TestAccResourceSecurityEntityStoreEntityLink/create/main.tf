variable "target_id" {
  description = "The target entity ID"
  type        = string
}

variable "entity_ids" {
  description = "The entity IDs to link"
  type        = list(string)
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_entity_store_entity_link" "test" {
  target_id  = var.target_id
  entity_ids = var.entity_ids
}
