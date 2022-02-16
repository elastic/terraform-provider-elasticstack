provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_dot_expander" "dot_expander" {
  field = "foo.bar"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "dot-expander-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_dot_expander.dot_expander.json
  ]
}
