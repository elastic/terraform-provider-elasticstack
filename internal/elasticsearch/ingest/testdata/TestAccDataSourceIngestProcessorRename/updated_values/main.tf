provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_rename" "test" {
  field        = "service.name"
  target_field = "service.type"
}
