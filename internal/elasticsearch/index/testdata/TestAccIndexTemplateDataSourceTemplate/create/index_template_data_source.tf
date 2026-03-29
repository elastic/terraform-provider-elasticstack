provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "mappings" {
  type = string
}

variable "settings" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]

  template {
    alias {
      name           = "my_alias"
      filter         = jsonencode({ term = { status = "active" } })
      index_routing  = "shard_1"
      search_routing = "shard_1"
      is_hidden      = false
      is_write_index = true
    }

    mappings = var.mappings
    settings = var.settings
  }
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
