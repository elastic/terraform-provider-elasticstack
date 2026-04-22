variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                      = var.index_name
  number_of_shards          = 1
  number_of_replicas        = 0
  refresh_interval          = "30s"
  max_result_window         = 5000
  max_ngram_diff            = 3
  gc_deletes                = "30s"
  blocks_read               = false
  blocks_write              = false
  routing_allocation_enable = "all"
  deletion_protection       = false
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = var.index_name
  depends_on = [elasticstack_elasticsearch_index.test]
}
