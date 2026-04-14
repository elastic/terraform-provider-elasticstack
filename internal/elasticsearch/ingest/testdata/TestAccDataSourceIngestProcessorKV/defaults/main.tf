provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_kv" "test" {
  field       = "event.original"
  field_split = "&"
  value_split = "="
}
