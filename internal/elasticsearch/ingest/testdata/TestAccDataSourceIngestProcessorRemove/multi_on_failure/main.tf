provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_remove" "test" {
  field          = ["user_agent"]
  ignore_missing = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "remove failed"
      }
    }),
    jsonencode({
      set = {
        field = "error.type"
        value = "remove"
      }
    }),
  ]
}
