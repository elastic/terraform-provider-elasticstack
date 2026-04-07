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

data "elasticstack_elasticsearch_info" "test_conn" {
  elasticsearch_connection {
    endpoints = [var.endpoint]
    ca_file   = var.ca_file
    cert_file = var.cert_file
    key_file  = var.key_file
  }
}
