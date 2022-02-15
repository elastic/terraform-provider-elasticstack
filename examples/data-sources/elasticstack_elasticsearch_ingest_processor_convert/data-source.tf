provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "convert" {
  description = "converts the content of the id field to an integer"
  field       = "id"
  type        = "integer"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "convert-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_convert.convert.json
  ]
}
