provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "convert" {
  field = "_ingest._value"
  type  = "integer"
}

data "elasticstack_elasticsearch_ingest_processor_foreach" "foreach" {
  field     = "values"
  processor = data.elasticstack_elasticsearch_ingest_processor_convert.convert.json
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "foreach-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_foreach.foreach.json
  ]
}
