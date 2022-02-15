provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_trim" "trim" {
  field = "foo"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "trim-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_trim.trim.json
  ]
}
