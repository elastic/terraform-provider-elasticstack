provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_json" "test" {
  field                = "document.json"
  allow_duplicate_keys = true
  description          = "Parse document JSON"
  if                   = "ctx.document?.json != null"
  ignore_failure       = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "json processor failed"
      }
    })
  ]
  tag = "json-tag"
}
