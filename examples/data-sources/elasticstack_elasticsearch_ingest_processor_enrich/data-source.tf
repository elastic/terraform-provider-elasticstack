provider "elasticstack" {
  elasticsearch {}
}

// the policy must exist before using this processor
// See example at: https://www.elastic.co/guide/en/elasticsearch/reference/current/match-enrich-policy-type.html
data "elasticstack_elasticsearch_ingest_processor_enrich" "enrich" {
  policy_name  = "users-policy"
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
