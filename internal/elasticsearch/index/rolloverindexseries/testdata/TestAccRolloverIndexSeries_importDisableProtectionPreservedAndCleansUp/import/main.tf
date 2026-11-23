variable "prefix" {
  type = string
}

variable "rollover_alias" {
  type = string
}

variable "deletion_protection" {
  type = bool
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_rollover_index_series" "test" {
  prefix              = var.prefix
  rollover_alias      = var.rollover_alias
  deletion_protection = var.deletion_protection
}
