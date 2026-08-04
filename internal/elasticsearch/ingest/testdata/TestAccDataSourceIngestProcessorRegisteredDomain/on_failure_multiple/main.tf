provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_registered_domain" "test" {
  field = "fqdn"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "registered domain failed"
      }
    }),
    jsonencode({
      remove = {
        field = "fqdn"
      }
    })
  ]
}
