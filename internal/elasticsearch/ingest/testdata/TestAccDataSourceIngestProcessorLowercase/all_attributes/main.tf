provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_lowercase" "test" {
  field          = "source_field"
  target_field   = "normalized_field"
  ignore_missing = true
  description    = "Normalize message to lowercase"
  if             = "ctx.source_field != null"
  ignore_failure = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "lowercase failed"
      }
    })
  ]
  tag = "lowercase-tag"
}
