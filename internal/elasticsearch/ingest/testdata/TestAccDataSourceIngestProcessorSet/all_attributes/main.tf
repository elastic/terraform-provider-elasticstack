provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_set" "test" {
  field          = "event.kind"
  value          = "alert"
  description    = "Set the event kind when a severity is present"
  if             = "ctx.severity != null"
  ignore_failure = true
  tag            = "set-event-kind"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "set processor failed"
      }
    })
  ]
}
