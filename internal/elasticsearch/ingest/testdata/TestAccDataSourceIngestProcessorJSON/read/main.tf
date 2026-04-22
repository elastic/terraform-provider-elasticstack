provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_json" "test" {
  field        = "string_source"
  target_field = "json_target"
}
