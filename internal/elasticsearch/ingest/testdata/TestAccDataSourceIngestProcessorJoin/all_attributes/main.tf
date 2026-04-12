provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_join" "test" {
  field          = "joined_array_field"
  separator      = "::"
  target_field   = "joined_field"
  description    = "Join array values into a single field"
  if             = "ctx.tags != null"
  ignore_failure = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "join failed"
      }
    })
  ]
  tag = "join-tags"
}
