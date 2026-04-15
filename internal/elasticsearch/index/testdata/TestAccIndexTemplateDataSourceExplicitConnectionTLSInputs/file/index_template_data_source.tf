provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "endpoint" {
  type = string
}

variable "ca_file" {
  type = string
}

variable "cert_file" {
  type = string
}

variable "key_file" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]
}

data "elasticstack_elasticsearch_index_template" "test_conn" {
  name = elasticstack_elasticsearch_index_template.test.name

  elasticsearch_connection {
    endpoints = [var.endpoint]
    ca_file   = var.ca_file
    cert_file = var.cert_file
    key_file  = var.key_file
  }
}
