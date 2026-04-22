variable "name" {
  type = string
}

variable "endpoint" {
  type = string
}

variable "bearer_token" {
  type      = string
  sensitive = true
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

data "elasticstack_elasticsearch_enrich_policy" "test" {
  name = elasticstack_elasticsearch_enrich_policy.policy.name

  elasticsearch_connection {
    endpoints    = [var.endpoint]
    bearer_token = var.bearer_token
  }
}
