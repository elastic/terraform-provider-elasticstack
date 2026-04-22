variable "name" {
  type = string
}

variable "endpoint" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name = var.name

  mappings = jsonencode({
    properties = {
      email      = { type = "keyword" }
      first_name = { type = "text" }
      last_name  = { type = "text" }
    }
  })
  deletion_protection = false
}

resource "elasticstack_elasticsearch_enrich_policy" "policy" {
  name          = var.name
  policy_type   = "match"
  indices       = [elasticstack_elasticsearch_index.my_index.name]
  match_field   = "email"
  enrich_fields = ["first_name", "last_name"]
  query         = jsonencode({ match_all = {} })
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "${var.name}-api-key"
  role_descriptors = jsonencode({
    enrich = {
      cluster = ["all"]
    }
  })
}

data "elasticstack_elasticsearch_enrich_policy" "test" {
  name = elasticstack_elasticsearch_enrich_policy.policy.name

  elasticsearch_connection {
    endpoints = [var.endpoint]
    api_key   = elasticstack_elasticsearch_security_api_key.test.encoded
    headers = {
      "X-Terraform-Test" = "enrich-policy"
      "X-Trace"          = "api-key"
    }
  }
}
