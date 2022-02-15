provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_circle" "circle" {
  field          = "circle"
  error_distance = 28.1
  shape_type     = "geo_shape"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "circle-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_circle.circle.json
  ]
}
