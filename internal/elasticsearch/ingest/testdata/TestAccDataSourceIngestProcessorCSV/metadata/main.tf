provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_csv" "test" {
  field         = "csv_payload"
  target_fields = ["first_name", "role"]
  description   = "Parse CSV when payload is present"
  if            = "ctx.csv_payload != null"
  tag           = "csv-tag"
}
