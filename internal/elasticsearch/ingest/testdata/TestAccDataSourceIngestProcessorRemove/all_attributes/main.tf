provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_remove" "test" {
  field = [
    "user_agent",
    "event.original",
  ]
  description    = "Remove user agent fields"
  if             = "ctx.user_agent != null"
  ignore_missing = true
  ignore_failure = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "remove failed"
      }
    })
  ]
  tag = "remove-fields"
}
