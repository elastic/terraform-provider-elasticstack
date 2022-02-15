provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_lowercase" "lowercase" {
  field = "foo"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "lowercase-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_lowercase.lowercase.json
  ]
}
