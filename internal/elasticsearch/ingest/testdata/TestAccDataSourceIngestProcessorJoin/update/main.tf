provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_join" "test" {
  field        = "updated_array_field"
  separator    = "|"
  target_field = "updated_joined_field"
  description  = "Join updated array values into a single field"
  if           = "ctx.updated_tags != null"
  on_failure = [
    jsonencode({
      set = {
        field = "error.code"
        value = "join_failed"
      }
    }),
    jsonencode({
      append = {
        field = "error.messages"
        value = "join failed again"
      }
    })
  ]
  tag = "join-updated-tag"
}
