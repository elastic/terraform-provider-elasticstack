provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_enrich" "test" {
  field        = "email"
  target_field = "user"
  policy_name  = "users-policy"

  on_failure = [
    "not-json"
  ]
}
