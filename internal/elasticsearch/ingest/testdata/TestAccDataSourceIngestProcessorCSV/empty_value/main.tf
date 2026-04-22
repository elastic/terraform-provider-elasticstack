provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_csv" "test" {
  field         = "csv_payload"
  target_fields = ["first_name", "role"]
  empty_value   = "N/A"
}
