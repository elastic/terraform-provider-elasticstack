provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_html_strip" "test_on_failure" {
  field = "body.html"

  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "html strip failed"
      }
    })
  ]
}
