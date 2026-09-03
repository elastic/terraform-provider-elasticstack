provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_convert" "test" {
  field = "my_field"
  type  = "auto"
}
