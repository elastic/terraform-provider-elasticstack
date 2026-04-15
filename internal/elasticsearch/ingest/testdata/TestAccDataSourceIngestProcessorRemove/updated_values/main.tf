provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_remove" "test" {
  field = [
    "host.name",
    "user.name",
  ]
}
