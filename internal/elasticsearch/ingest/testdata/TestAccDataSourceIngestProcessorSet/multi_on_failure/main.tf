provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_set" "test" {
  field = "count"
  value = 1
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "set processor failed"
      }
    }),
    jsonencode({
      set = {
        field = "error.type"
        value = "set"
      }
    })
  ]
}
