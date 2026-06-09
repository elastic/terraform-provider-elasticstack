provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_circle" "test" {
  field          = "circle"
  error_distance = 28.1
  shape_type     = "geo_shape"

  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "circle failed"
      }
    }),
    jsonencode({
      set = {
        field = "error.type"
        value = "circle"
      }
    })
  ]
}
