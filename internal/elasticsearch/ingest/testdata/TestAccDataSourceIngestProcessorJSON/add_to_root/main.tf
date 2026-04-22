provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_json" "test" {
  field                         = "json_payload"
  add_to_root                   = true
  add_to_root_conflict_strategy = "merge"
}
