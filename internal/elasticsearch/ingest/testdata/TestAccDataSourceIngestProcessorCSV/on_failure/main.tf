provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_csv" "test" {
  field          = "csv_payload"
  target_fields  = ["first_name", "role"]
  ignore_failure = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "csv failed"
      }
    })
  ]
}
