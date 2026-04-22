variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name = var.index_name

  mappings = jsonencode({
    properties = {
      sort_key = { type = "keyword" }
    }
  })

  number_of_shards                  = 2
  number_of_routing_shards          = 2
  routing_partition_size            = 1
  load_fixed_bitset_filters_eagerly = true
  shard_check_on_startup            = "false"
  sort_field                        = ["sort_key"]
  sort_order                        = ["asc"]
  deletion_protection               = false
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = var.index_name
  depends_on = [elasticstack_elasticsearch_index.test]
}
