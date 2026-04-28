provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "data_retention" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]

  data_stream {}

  template {
    lifecycle {
      data_retention = var.data_retention
    }
  }
}
