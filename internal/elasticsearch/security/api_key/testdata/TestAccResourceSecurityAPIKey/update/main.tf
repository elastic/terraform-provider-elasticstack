variable "api_key_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = var.api_key_name

  role_descriptors = jsonencode({
    role-a = {
      cluster = ["manage"]
      indices = [{
        names                    = ["index-b*"]
        privileges               = ["read"]
        allow_restricted_indices = false
      }]
    }
  })

  expiration = "1d"
}
