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

data "elasticstack_elasticsearch_info" "test_conn" {
  elasticsearch_connection {
    endpoints = [var.endpoint]
    ca_data   = var.ca_data
    cert_data = var.cert_data
    key_data  = var.key_data
  }
}
