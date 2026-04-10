provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_sort" "test" {
  field = "array_field_to_sort"
  order = "desc"
}
