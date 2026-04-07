provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_join" "test" {
  field     = "joined_array_field"
  separator = "-"
}
