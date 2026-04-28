provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name = var.template_name

  priority = 100
  index_patterns = [
    "${var.template_name}-logs-*",
    "${var.template_name}-metrics-*",
    "${var.template_name}-traces-*",
  ]
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
