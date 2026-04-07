variable "endpoints" {
  type = list(string)
}

variable "headers" {
  type = map(string)
}

variable "username" {
  type = string
}

variable "password" {
  type      = string
  sensitive = true
}

provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_info" "test_conn" {
  elasticsearch_connection {
    endpoints = var.endpoints
    headers   = var.headers
    username  = var.username
    password  = var.password
  }
}
