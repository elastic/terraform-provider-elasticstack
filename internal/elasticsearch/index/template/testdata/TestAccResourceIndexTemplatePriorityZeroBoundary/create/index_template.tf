provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "priority" {
  type = number
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]
  priority       = var.priority
}
