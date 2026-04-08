provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_dissect" "test_on_failure" {
  field   = "message"
  pattern = "%{clientip} %{ident} %{auth}"

  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "{{ _ingest.on_failure_message }}"
      }
    })
  ]
}
