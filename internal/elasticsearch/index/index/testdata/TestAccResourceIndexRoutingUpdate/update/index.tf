variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name = var.index_name

  routing_allocation_enable = "all"
  routing_rebalance_enable  = "replicas"

  deletion_protection = false
}
