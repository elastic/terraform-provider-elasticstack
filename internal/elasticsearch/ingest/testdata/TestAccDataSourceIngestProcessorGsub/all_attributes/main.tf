provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_gsub" "test" {
  field          = "field1"
  pattern        = "\\."
  replacement    = "-"
  target_field   = "normalized_field"
  ignore_missing = true
  description    = "Normalize a dotted field"
  if             = "ctx.message != null"
  ignore_failure = true
  on_failure = [
    jsonencode({
      append = {
        field = "errors"
        value = ["gsub failed"]
      }
    })
  ]
  tag = "gsub-normalize"
}
