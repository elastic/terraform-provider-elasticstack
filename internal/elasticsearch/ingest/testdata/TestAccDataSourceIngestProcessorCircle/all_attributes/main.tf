provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_circle" "test" {
  field          = "location"
  target_field   = "location_shape"
  ignore_missing = true
  error_distance = 5
  shape_type     = "geo_shape"
  description    = "Convert circle to polygon"
  if             = "ctx.location != null"
  ignore_failure = true
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "circle failed"
      }
    })
  ]
  tag = "circle-tag"
}
