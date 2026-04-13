variable "endpoints" {
  type = list(string)
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

data "elasticstack_elasticsearch_indices" "test_conn" {
  target = ".security-*"

  elasticsearch_connection {
    username  = var.username
    password  = var.password
    endpoints = var.endpoints
    insecure  = true
  }
}
