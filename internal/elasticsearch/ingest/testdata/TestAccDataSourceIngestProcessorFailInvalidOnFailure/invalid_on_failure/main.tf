provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_fail" "test" {
  message = "Reject documents without a service name"

  on_failure = [
    "not-json"
  ]
}
