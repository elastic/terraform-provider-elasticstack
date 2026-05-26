variable "api_key_name" {
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

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = var.api_key_name

  role_descriptors = jsonencode({
    role-a = {
      cluster = ["all"]
      indices = [{
        names                    = ["index-a*"]
        privileges               = ["read"]
        allow_restricted_indices = false
      }]
    }
  })

  elasticsearch_connection {
    endpoints = var.endpoints
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.username != "" ? var.username : null
    password  = var.password != "" ? var.password : null
    insecure  = true
  }
}
