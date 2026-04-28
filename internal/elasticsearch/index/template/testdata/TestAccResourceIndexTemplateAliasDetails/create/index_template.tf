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
      routing        = "route_common_v1"
      search_routing = "search_explicit_v1"
      index_routing  = "index_explicit_v1"
    }
  }
}
