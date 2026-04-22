provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_append" "tags" {
  field = "tags"
  value = ["production", "{{{app}}}", "{{{owner}}}"]
}

resource "elasticstack_elasticsearch_ingest_pipeline" "pipeline_a" {
  name = "pipeline_a"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_append.tags.json
  ]
}

data "elasticstack_elasticsearch_ingest_processor_pipeline" "test" {
  name = elasticstack_elasticsearch_ingest_pipeline.pipeline_a.name
}
