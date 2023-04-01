provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name = "my-index"

  mappings = jsonencode({
    properties = {
      email      = { type = "text" }
      first_name = { type = "text" }
      last_name  = { type = "text" }
    }
  })
  deletion_protection = false
}

resource "elasticstack_elasticsearch_enrich_policy" "policy1" {
  name          = "policy1"
  policy_type   = "match"
  indices       = [elasticstack_elasticsearch_index.my_index.name]
  match_field   = "email"
  enrich_fields = ["first_name", "last_name"]
  query = jsonencode({
    bool = {
      must     = [{ term = { b = "A" } }]
      must_not = [{ term = { a = "B" } }]
    }
  })
}

data "elasticstack_elasticsearch_enrich_policy" "policy" {
  name = "policy1"
}

output "name" {
  value = data.elasticstack_elasticsearch_enrich_policy.policy.name
}
output "match_field" {
  value = data.elasticstack_elasticsearch_enrich_policy.policy.match_field
}
output "indices" {
  value = data.elasticstack_elasticsearch_enrich_policy.policy.indices
}
output "policy_type" {
  value = data.elasticstack_elasticsearch_enrich_policy.policy.policy_type
}
output "enrich_fields" {
  value = data.elasticstack_elasticsearch_enrich_policy.policy.enrich_fields
}
output "query" {
  value = jsondecode(data.elasticstack_elasticsearch_enrich_policy.policy.query)
}