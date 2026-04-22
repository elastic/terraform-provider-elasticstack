provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_community_id" "test_on_failure" {
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "community id failed"
      }
    })
  ]
}
