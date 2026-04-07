variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.name
  index_patterns = ["${var.name}-*"]

  data_stream {}

  template {
    alias {
      name           = "detailed_alias_initial"
      is_hidden      = true
      is_write_index = true
      routing        = "shard_1"
      search_routing = "shard_1"
      index_routing  = "shard_1"
    }

    lifecycle {
      data_retention = "30d"
    }
  }
}
