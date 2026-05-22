variable "api_key_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = var.api_key_name

  role_descriptors = jsonencode({
    role-default = {
      cluster = ["monitor"]
      indices = [{
        names      = ["logs-*", "metrics-*"]
        privileges = ["read", "view_index_metadata"]
        # Note: allow_restricted_indices is NOT specified here - should default to false
      }]
    }
  })

  expiration = "2d"
}
