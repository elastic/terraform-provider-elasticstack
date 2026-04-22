provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "alias_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]

  template {
    alias {
      name = var.alias_name
    }
  }
}
