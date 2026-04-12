variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "index_a" {
  name = "${var.name}-a"

  mappings = jsonencode({
    properties = {
      email      = { type = "keyword" }
      first_name = { type = "text" }
    }
  })
  deletion_protection = false
}

resource "elasticstack_elasticsearch_enrich_policy" "policy" {
  name          = var.name
  policy_type   = "match"
  indices       = [elasticstack_elasticsearch_index.index_a.name]
  match_field   = "email"
  enrich_fields = ["first_name"]
}

data "elasticstack_elasticsearch_enrich_policy" "test" {
  name = elasticstack_elasticsearch_enrich_policy.policy.name
}
