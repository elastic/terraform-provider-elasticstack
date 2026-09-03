provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "inner" {
  field = "_ingest._value"
  type  = "integer"
}

data "elasticstack_elasticsearch_ingest_processor_foreach" "test" {
  field          = "values"
  processor      = data.elasticstack_elasticsearch_ingest_processor_convert.inner.json
  ignore_missing = true
  description    = "foreach test"
  if             = "ctx.values != null"
  tag            = "foreach-tag"
  ignore_failure = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "foreach failed"
      }
    })
  ]
}
