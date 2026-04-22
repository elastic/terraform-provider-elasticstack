provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_urldecode" "test_on_failure" {
  field = "source.url"

  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "{{ _ingest.on_failure_message }}"
      }
    })
  ]
}
