provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_lowercase" "test" {
  field = "foo"
}
