variable "prefix" {
  type = string
}

variable "rollover_alias" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_rollover_index_series" "test" {
  prefix         = var.prefix
  rollover_alias = var.rollover_alias
}
