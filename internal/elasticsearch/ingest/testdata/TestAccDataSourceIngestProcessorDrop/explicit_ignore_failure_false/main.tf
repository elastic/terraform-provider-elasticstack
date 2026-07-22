provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_drop" "test" {
  ignore_failure = false
}
