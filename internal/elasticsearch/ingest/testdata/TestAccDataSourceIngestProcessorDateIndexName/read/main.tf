provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_date_index_name" "test" {
  description       = "monthly date-time index naming"
  field             = "date1"
  index_name_prefix = "my-index-"
  date_rounding     = "M"
}
