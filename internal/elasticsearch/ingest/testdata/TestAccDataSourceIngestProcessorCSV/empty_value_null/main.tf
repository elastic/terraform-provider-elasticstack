provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_csv" "test" {
  field         = "csv_payload"
  target_fields = ["col1", "col2"]
  empty_value   = "NULL"
}
