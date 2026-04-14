provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]

  data_stream {}

  template {
    alias {
      name           = "detailed_alias_initial"
      filter         = jsonencode({ term = { status = "active" } })
      is_hidden      = true
      is_write_index = true
      search_routing = "shard_1"
      index_routing  = "shard_1"
    }

    lifecycle {
      data_retention = "30d"
    }
  }
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
