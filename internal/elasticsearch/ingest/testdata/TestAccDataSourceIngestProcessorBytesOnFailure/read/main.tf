provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_bytes" "test_on_failure" {
  field = "file.size"

  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "{{ _ingest.on_failure_message }}"
      }
    })
  ]
}
