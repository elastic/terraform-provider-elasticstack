provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_date_index_name" "test" {
  field         = "timestamp_raw"
  date_rounding = "d"
}
