provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "mappings" {
  # Expected to be a JSON-encoded string.
  type = string
}

variable "settings" {
  # Expected to be a JSON-encoded string.
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]

  template {
    alias {
      name           = "my_alias_v2"
      filter         = jsonencode({ bool = { must = [{ term = { "service.name" = "api" } }, { term = { status = "active" } }] } })
      index_routing  = "shard_2"
      search_routing = "shard_2"
      is_hidden      = true
      is_write_index = false
    }

    mappings = var.mappings
    settings = var.settings
  }
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
