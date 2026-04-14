variable "index_name" {
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

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  deletion_protection = false
}

data "elasticstack_elasticsearch_indices" "test_conn" {
  target     = elasticstack_elasticsearch_index.test.name
  depends_on = [elasticstack_elasticsearch_index.test]

  elasticsearch_connection {
    endpoints = [var.endpoint]
    ca_file   = var.ca_file
    cert_file = var.cert_file
    key_file  = var.key_file
  }
}
