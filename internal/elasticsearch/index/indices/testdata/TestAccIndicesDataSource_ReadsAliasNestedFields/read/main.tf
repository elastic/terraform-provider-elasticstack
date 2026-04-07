variable "index_name" {
  type = string
}

variable "alias_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  number_of_shards    = 1
  number_of_replicas  = 0
  deletion_protection = false

  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index_alias" "test" {
  name = var.alias_name

  write_index = {
    name           = elasticstack_elasticsearch_index.test.name
    filter         = jsonencode({ term = { "status" = "active" } })
    index_routing  = "shard-1"
    is_hidden      = false
    search_routing = "shard-1"
  }
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = var.index_name
  depends_on = [elasticstack_elasticsearch_index_alias.test]
}
