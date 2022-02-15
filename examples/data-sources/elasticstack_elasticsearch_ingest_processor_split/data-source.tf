provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_split" "split" {
  field     = "my_field"
  separator = "\\s+"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "split-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_split.split.json
  ]
}
