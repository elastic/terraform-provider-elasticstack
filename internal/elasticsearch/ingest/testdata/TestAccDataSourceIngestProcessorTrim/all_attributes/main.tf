provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_trim" "test" {
  field          = "message"
  target_field   = "trimmed_message"
  ignore_missing = true
  ignore_failure = true
  description    = "Trim whitespace from message"
  if             = "ctx.message != null"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "trim failed"
      }
    })
  ]
  tag = "trim-message"
}
