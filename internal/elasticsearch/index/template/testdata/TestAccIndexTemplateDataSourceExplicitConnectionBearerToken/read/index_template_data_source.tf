provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "endpoint" {
  type = string
}

variable "bearer_token" {
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
    endpoints                = [var.endpoint]
    bearer_token             = var.bearer_token
    es_client_authentication = "Authorization"
  }
}
