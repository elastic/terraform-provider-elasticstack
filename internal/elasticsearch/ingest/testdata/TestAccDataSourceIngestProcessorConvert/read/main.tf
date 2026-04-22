provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "test" {
  description = "converts the content of the id field to an integer"
  field       = "id"
  type        = "integer"
}
