provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_uppercase" "uppercase" {
  field = "foo"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "uppercase-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_uppercase.uppercase.json
  ]
}
