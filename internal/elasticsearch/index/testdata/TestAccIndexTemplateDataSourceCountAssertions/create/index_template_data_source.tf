provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "component_1" {
  type = string
}

variable "component_2" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-a-*", "${var.template_name}-b-*"]

  composed_of                        = [var.component_1, var.component_2]
  ignore_missing_component_templates = [var.component_1, var.component_2]
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
