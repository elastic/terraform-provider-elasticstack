provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_component_template" "component_a" {
  name = "${var.template_name}-logscomponent-a@custom"

  template {
    settings = jsonencode({
      number_of_shards = "1"
    })
  }
}

resource "elasticstack_elasticsearch_component_template" "component_b" {
  name = "${var.template_name}-logscomponent-b@custom"

  template {
    settings = jsonencode({
      number_of_shards = "1"
    })
  }
}

resource "elasticstack_elasticsearch_component_template" "component_c" {
  name = "${var.template_name}-logscomponent-c@custom"

  template {
    settings = jsonencode({
      number_of_shards = "1"
    })
  }
}

resource "elasticstack_elasticsearch_index_template" "test_component" {
  name = var.template_name

  index_patterns = ["${var.template_name}-logscomponent-*"]
  composed_of = [
    elasticstack_elasticsearch_component_template.component_a.name,
    elasticstack_elasticsearch_component_template.component_c.name,
  ]

  ignore_missing_component_templates = [
    elasticstack_elasticsearch_component_template.component_c.name,
  ]

  template {}
}
