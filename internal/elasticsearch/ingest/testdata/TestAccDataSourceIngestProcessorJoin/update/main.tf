provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_join" "test" {
  field     = "updated_array_field"
  separator = "|"
}
