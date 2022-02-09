provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_append" "tags" {
  field = "tags"
  value = ["production", "{{{app}}}", "{{{owner}}}"]

}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "append-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_append.tags.json
  ]
}
