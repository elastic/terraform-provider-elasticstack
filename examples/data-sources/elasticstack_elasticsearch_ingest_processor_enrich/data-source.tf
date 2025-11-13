provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name = "example-enrich-ingest-processor-index"

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
  name          = "example-enrich-ingest-processor-policy"
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

// the policy must exist before using this processor
// See example at: https://www.elastic.co/guide/en/elasticsearch/reference/current/match-enrich-policy-type.html
data "elasticstack_elasticsearch_ingest_processor_enrich" "enrich" {
  policy_name  = elasticstack_elasticsearch_enrich_policy.policy1.name
  field        = "email"
  target_field = "user"
  max_matches  = 1
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "enrich-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_enrich.enrich.json
  ]
}
