provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "allow_custom_routing" {
  type = bool
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]

  data_stream {
    allow_custom_routing = var.allow_custom_routing
  }
}
