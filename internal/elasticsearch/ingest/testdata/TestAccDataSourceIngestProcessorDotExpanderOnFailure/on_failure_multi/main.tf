provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_dot_expander" "test" {
  field = "foo.bar"

  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "{{ _ingest.on_failure_message }}"
      }
    }),
    jsonencode({
      set = {
        field = "error.type"
        value = "dot_expander"
      }
    })
  ]
}
