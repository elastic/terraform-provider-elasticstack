provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_set" "set" {
  field = "count"
  value = 1
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "set-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_set.set.json
  ]
}
