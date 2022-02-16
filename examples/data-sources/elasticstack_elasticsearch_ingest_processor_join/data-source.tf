provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_join" "join" {
  field     = "joined_array_field"
  separator = "-"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "join-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_join.join.json
  ]
}
