provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_uri_parts" "test" {
  field = "input_field"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "uri parts failed"
      }
    }),
    jsonencode({
      set = {
        field = "error.type"
        value = "uri_parts"
      }
    })
  ]
}
