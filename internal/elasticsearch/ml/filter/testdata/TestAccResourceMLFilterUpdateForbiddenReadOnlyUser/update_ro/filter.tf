variable "filter_id" {
  type = string
}

variable "endpoints" {
  type = list(string)
}

variable "ro_username" {
  type = string
}

variable "ro_password" {
  type      = string
  sensitive = true
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_filter" "test" {
  filter_id   = var.filter_id
  description = "Read-only user cannot apply this update"
  items       = ["*.example.com", "denied-addition.example.org"]

  elasticsearch_connection {
    endpoints = var.endpoints
    username  = var.ro_username
    password  = var.ro_password
    insecure  = true
  }
}
