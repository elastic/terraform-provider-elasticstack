variable "api_key_name" {
  type = string
}

variable "endpoints" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test_connection" {
  name = var.api_key_name

  role_descriptors = jsonencode({
    role-a = {
      cluster = ["all"]
      indices = [{
        names                    = ["*"]
        privileges               = ["all"]
        allow_restricted_indices = false
      }]
    }
  })

  expiration = "1d"
}

data "elasticstack_elasticsearch_security_user" "test" {
  username = "elastic"

  elasticsearch_connection {
    endpoints = [var.endpoints]
    api_key   = elasticstack_elasticsearch_security_api_key.test_connection.encoded
  }
}
