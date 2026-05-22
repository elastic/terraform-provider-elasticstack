variable "api_key_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

ephemeral "elasticstack_elasticsearch_security_api_key" "test" {
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

  invalidate_on_close = false
}

provider "echo" {
  data = ephemeral.elasticstack_elasticsearch_security_api_key.test
}

resource "echo" "capture" {}
