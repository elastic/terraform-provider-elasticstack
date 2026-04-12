provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_json" "test" {
  field        = "updated_string_source"
  target_field = "updated_json_target"
}
