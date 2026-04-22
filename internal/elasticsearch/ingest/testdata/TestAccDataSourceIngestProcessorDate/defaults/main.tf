provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_date" "test" {
  field   = "timestamp_raw"
  formats = ["ISO8601"]
}
