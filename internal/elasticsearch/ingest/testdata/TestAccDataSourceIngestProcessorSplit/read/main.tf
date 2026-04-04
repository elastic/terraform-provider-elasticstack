provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_split" "test" {
  field     = "my_field"
  separator = "\\s+"
}
