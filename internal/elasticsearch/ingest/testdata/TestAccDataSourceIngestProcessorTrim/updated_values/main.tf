provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_trim" "test" {
  field        = "my.field"
  target_field = "trimmed_field"
}
