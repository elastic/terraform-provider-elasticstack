provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_bytes" "test" {
  field = "file.size"
}
