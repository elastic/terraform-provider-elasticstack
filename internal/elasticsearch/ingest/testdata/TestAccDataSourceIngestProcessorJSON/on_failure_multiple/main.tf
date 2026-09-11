provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_json" "test" {
  field = "document.json"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "json processor failed"
      }
    }),
    jsonencode({
      set = {
        field = "error.tag"
        value = "json-parse-error"
      }
    })
  ]
}
