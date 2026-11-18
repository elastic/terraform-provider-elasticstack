provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_script" "test" {
  source = "ctx.result = 'ok';"

  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "script processor failed"
      }
    }),
    jsonencode({
      set = {
        field = "error.type"
        value = "script_error"
      }
    }),
  ]
}
