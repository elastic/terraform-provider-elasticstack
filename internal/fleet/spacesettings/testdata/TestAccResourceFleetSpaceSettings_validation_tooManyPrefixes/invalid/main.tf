variable "space_id" {
  type = string
}

variable "allowed_namespace_prefixes" {
  type = set(string)
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_space_settings" "test" {
  space_id                   = var.space_id
  allowed_namespace_prefixes = var.allowed_namespace_prefixes
}
