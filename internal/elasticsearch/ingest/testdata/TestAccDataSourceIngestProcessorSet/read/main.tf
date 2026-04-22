provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_set" "test" {
  field = "count"
  value = 1
}
