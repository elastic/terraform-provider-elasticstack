provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test_component" {
  name = var.template_name

  index_patterns = ["tf-acc-component-${var.template_name}-*"]
  composed_of    = ["${var.template_name}-logscomponent@custom"]

  ignore_missing_component_templates = ["${var.template_name}-logscomponent@custom"]
}

data "elasticstack_elasticsearch_index_template" "test_component" {
  name = elasticstack_elasticsearch_index_template.test_component.name
}
