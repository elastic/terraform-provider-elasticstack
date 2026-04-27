provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "hidden" {
  type = bool
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]

  data_stream {
    hidden = var.hidden
  }
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
