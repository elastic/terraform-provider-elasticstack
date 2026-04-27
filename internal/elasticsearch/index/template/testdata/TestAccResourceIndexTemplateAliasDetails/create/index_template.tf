provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]

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
