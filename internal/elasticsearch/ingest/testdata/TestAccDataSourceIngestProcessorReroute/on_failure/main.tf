provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_reroute" "test" {
  destination    = "logs-fallback-default"
  ignore_failure = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "reroute failed"
      }
    })
  ]
}
