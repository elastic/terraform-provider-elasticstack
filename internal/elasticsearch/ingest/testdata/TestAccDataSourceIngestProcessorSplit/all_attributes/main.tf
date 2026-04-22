provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_split" "test" {
  field             = "message"
  separator         = ","
  target_field      = "message_parts"
  preserve_trailing = true
  ignore_missing    = true
  description       = "Split a comma-delimited message"
  if                = "ctx.message != null"
  ignore_failure    = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "split failed"
      }
    })
  ]
  tag = "split-message"
}
