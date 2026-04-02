provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
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
  type = string
}

variable "password" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_conn" {
  name = var.policy_name

  elasticsearch_connection {
    endpoints = var.endpoints
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
    insecure  = true
  }

  hot {
    rollover {
      max_age = "7d"
    }
  }

  delete {
    delete {}
  }
}
