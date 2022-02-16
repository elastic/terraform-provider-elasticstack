provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_sort" "sort" {
  field = "array_field_to_sort"
  order = "desc"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "sort-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_sort.sort.json
  ]
}
