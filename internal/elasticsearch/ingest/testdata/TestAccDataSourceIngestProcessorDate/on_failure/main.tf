provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_date" "test" {
  field   = "initial_date"
  formats = ["dd/MM/yyyy HH:mm:ss"]
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "date parse failed"
      }
    })
  ]
}
