provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_lowercase" "test" {
  field        = "updated_source_field"
  target_field = "updated_normalized_field"
}
