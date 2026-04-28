provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "endpoint" {
  type = string
}

variable "ca_data" {
  type = string
}

variable "cert_data" {
  type = string
}

variable "key_data" {
  type      = string
  sensitive = true
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]
}

data "elasticstack_elasticsearch_index_template" "test_conn" {
  name = elasticstack_elasticsearch_index_template.test.name

  elasticsearch_connection {
    endpoints = [var.endpoint]
    ca_data   = var.ca_data
    cert_data = var.cert_data
    key_data  = var.key_data
  }
}
