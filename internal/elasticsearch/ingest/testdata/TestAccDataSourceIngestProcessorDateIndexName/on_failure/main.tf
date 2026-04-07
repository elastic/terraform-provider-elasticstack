provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_date_index_name" "test" {
  field         = "date1"
  date_rounding = "M"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "date index routing failed"
      }
    })
  ]
}
