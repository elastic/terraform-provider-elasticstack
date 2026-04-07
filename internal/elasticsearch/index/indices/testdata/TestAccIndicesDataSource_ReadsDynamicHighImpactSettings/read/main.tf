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
      field1 = { type = "text" }
    }
  })

  number_of_shards              = 1
  number_of_replicas            = 0
  auto_expand_replicas          = "0-5"
  search_idle_after             = "30s"
  query_default_field           = ["field1"]
  blocks_read_only              = false
  blocks_read_only_allow_delete = false
  deletion_protection           = false
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = var.index_name
  depends_on = [elasticstack_elasticsearch_index.test]
}
