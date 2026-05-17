provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "allow_auto_create" {
  type = bool
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name              = var.template_name
  index_patterns    = ["${var.template_name}-*"]
  allow_auto_create = var.allow_auto_create
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
