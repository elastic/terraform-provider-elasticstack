provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "metadata" {
  type = string
}

// Terraform reserves "version" as a variable name, so this fixture uses template_version.
variable "template_version" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]
  version        = tonumber(var.template_version)
  metadata       = var.metadata
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
