variable "prefix" {
  type = string
}

variable "index_a" {
  type = string
}

variable "index_b" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "a" {
  name                = var.index_a
  number_of_shards    = 1
  number_of_replicas  = 0
  deletion_protection = false
}

resource "elasticstack_elasticsearch_index" "b" {
  name                = var.index_b
  number_of_shards    = 1
  number_of_replicas  = 0
  deletion_protection = false
}

data "elasticstack_elasticsearch_indices" "wildcard" {
  target     = "${var.prefix}*"
  depends_on = [elasticstack_elasticsearch_index.a, elasticstack_elasticsearch_index.b]
}

data "elasticstack_elasticsearch_indices" "exact" {
  target     = var.index_a
  depends_on = [elasticstack_elasticsearch_index.a, elasticstack_elasticsearch_index.b]
}
