provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_append" "test" {
  field = "f"
  value = []
}
