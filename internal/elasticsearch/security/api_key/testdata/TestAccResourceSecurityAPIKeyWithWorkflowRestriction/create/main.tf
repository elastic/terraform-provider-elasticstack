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
      cluster = ["all"]
      indices = [{
        names                    = ["index-a*"]
        privileges               = ["read"]
        allow_restricted_indices = false
      }]
      restriction = {
        workflows = ["search_application_query"]
      }
    }
  })

  expiration = "1d"
}
