provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "metadata" {
  type = string
}

variable "mappings" {
  type = string
}

variable "template_version" {
  type = number
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]
  version        = var.template_version
  metadata       = var.metadata

  template {
    mappings = var.mappings
  }
}
