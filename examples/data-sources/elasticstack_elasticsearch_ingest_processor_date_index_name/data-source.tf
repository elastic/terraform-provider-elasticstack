provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_date_index_name" "date_index_name" {
  description       = "monthly date-time index naming"
  field             = "date1"
  index_name_prefix = "my-index-"
  date_rounding     = "M"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "date-index-name-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_date_index_name.date_index_name.json
  ]
}
