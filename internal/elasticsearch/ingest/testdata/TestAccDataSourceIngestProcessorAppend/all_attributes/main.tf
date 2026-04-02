provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_append" "test" {
  description      = "Append a numeric-like error code to tags"
  field            = "tags"
  value            = ["404"]
  allow_duplicates = false
  media_type       = "application/json"
  if               = "ctx.error != null"
  ignore_failure   = true
  tag              = "append-tag"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "append failed"
      }
    })
  ]
}
