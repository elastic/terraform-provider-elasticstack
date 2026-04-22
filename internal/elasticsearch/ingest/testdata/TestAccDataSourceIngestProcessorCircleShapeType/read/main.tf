provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_circle" "test_shape" {
  field          = "circle"
  error_distance = 10
  shape_type     = "shape"
}
