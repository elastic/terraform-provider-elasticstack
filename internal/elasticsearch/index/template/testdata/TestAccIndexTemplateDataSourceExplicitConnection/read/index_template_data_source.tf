provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "endpoint" {
  type = string
}

variable "username" {
  type = string
}

variable "password" {
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
    username  = var.username
    password  = var.password
    headers = {
      XTerraformTest = "basic-auth"
    }
    insecure = true
  }
}
