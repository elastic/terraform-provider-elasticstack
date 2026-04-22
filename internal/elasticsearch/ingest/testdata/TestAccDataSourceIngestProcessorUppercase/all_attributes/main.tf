provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_uppercase" "test" {
  field          = "source_field"
  target_field   = "uppercased_field"
  ignore_missing = true
  description    = "Normalize message to uppercase"
  if             = "ctx.source_field != null"
  ignore_failure = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "uppercase failed"
      }
    })
  ]
  tag = "uppercase-tag"
}
