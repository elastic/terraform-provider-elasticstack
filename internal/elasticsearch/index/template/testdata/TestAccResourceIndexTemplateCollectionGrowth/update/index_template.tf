provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_component_template" "component_a" {
  name = "${var.template_name}-comp-a@custom"

  template {
    settings = jsonencode({
      number_of_shards = "1"
    })
  }
}

resource "elasticstack_elasticsearch_component_template" "component_b" {
  name = "${var.template_name}-comp-b@custom"

  template {
    settings = jsonencode({
      number_of_shards = "1"
    })
  }
}

resource "elasticstack_elasticsearch_component_template" "component_c" {
  name = "${var.template_name}-comp-c@custom"

  template {
    settings = jsonencode({
      number_of_shards = "1"
    })
  }
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]

  composed_of                        = [elasticstack_elasticsearch_component_template.component_a.name]
  ignore_missing_component_templates = [elasticstack_elasticsearch_component_template.component_b.name]
}
