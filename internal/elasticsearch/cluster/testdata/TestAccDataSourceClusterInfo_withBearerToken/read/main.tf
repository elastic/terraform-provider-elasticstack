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

data "elasticstack_elasticsearch_info" "test_conn" {
  elasticsearch_connection {
    endpoints                = [var.endpoint]
    bearer_token             = var.bearer_token
    es_client_authentication = "Authorization"
  }
}
