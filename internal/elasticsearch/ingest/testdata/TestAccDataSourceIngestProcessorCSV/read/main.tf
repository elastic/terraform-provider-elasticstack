provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_csv" "test" {
  field         = "my_field"
  target_fields = ["field1", "field2"]
}
