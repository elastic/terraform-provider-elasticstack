provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_fingerprint" "test" {
  fields = ["user"]
}
