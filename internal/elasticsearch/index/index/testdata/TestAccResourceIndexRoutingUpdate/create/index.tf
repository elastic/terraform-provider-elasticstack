variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name = var.index_name

  routing_allocation_enable = "primaries"
  routing_rebalance_enable  = "primaries"

  deletion_protection = false
}
