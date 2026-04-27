provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name = var.template_name

  priority       = 100
  index_patterns = ["tf-acc-${var.template_name}-*"]
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
