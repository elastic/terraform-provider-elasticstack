provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "test" {
  field = "id"
  type  = "integer"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "convert failed"
      }
    })
  ]
}
