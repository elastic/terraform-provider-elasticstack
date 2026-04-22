provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_circle" "test" {
  field          = "circle"
  error_distance = 28.1
  shape_type     = "geo_shape"
}
