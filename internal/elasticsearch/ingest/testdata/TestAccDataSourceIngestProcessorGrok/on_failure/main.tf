provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_grok" "test" {
  field    = "message"
  patterns = ["%%{WORD:log.level}: %%{GREEDYDATA:message}"]
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "failed"
      }
    })
  ]
}
