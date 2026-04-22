provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_gsub" "test" {
  field          = "field2"
  pattern        = ":"
  replacement    = "_"
  target_field   = "normalized_field"
  ignore_missing = true
  description    = "Normalize colon-delimited field"
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
