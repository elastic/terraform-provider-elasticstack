provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test_component" {
  name = var.template_name

  index_patterns = ["${var.template_name}-logscomponent-*"]
  composed_of    = ["${var.template_name}-logscomponent-updated@custom"]

  ignore_missing_component_templates = ["${var.template_name}-logscomponent-updated@custom"]

  template {}
}
