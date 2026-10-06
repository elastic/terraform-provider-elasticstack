provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = var.template_name
}
