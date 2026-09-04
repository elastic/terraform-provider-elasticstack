provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_reroute" "test" {
  destination    = "logs-multi-default"
  ignore_failure = true
  tag            = "reroute-multi-failure"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "reroute failed"
      }
    }),
    jsonencode({
      remove = {
        field = "error.stack_trace"
      }
    })
  ]
}
