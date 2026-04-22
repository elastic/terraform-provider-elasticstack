variable "name" {
  type = string
}

variable "endpoints" {
  type = list(string)
}

variable "api_key" {
  type    = string
  default = ""
}

variable "username" {
  type    = string
  default = ""
}

variable "password" {
  type    = string
  default = ""
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.name
  index_patterns = ["${var.name}-connection-*"]

  elasticsearch_connection {
    endpoints = var.endpoints
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.username != "" ? var.username : null
    password  = var.password != "" ? var.password : null
    insecure  = true
  }

  template {
    settings = jsonencode({
      number_of_shards = "1"
    })
  }
}
