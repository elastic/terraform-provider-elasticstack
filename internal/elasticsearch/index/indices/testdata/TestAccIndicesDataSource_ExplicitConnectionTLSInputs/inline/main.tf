variable "index_name" {
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
    ca_data   = var.ca_data
    cert_data = var.cert_data
    key_data  = var.key_data
  }
}
