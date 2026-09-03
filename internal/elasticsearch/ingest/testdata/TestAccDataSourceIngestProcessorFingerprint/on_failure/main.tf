provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_fingerprint" "test" {
  fields = ["user"]

  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "fingerprint failed"
      }
    }),
    jsonencode({
      set = {
        field = "error.type"
        value = "fingerprint"
      }
    })
  ]
}
