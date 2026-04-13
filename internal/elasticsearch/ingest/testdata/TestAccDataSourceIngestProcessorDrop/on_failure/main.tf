provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_drop" "test" {
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "drop failed"
      }
    })
  ]
}
