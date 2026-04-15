variable "index_name" {
  type = string
}

variable "endpoint" {
  type = string
}

variable "bearer_token" {
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
    endpoints                = [var.endpoint]
    bearer_token             = var.bearer_token
    es_client_authentication = "Authorization"
  }
}
