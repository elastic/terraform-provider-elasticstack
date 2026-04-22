provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "test" {
  field = "_ingest._value"
  type  = "integer"
}

data "elasticstack_elasticsearch_ingest_processor_foreach" "test" {
  field     = "values"
  processor = data.elasticstack_elasticsearch_ingest_processor_convert.test.json
}
