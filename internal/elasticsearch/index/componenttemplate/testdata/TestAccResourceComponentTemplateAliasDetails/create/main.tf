variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "test" {
  name = var.name

  template {
    alias {
      name           = "detailed_alias"
      is_hidden      = true
      is_write_index = true
      routing        = "shard_1"
      search_routing = "shard_1"
      index_routing  = "shard_1"
    }
  }
}
