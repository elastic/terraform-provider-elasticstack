provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_set" "test" {
  field     = "archived_count"
  copy_from = "count"
}
