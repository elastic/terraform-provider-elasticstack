provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_uppercase" "test" {
  field        = "updated_source_field"
  target_field = "updated_uppercased_field"
}
