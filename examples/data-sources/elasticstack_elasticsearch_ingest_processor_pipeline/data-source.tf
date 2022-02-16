provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_append" "append_tags" {
  field = "tags"
  value = ["production", "{{{app}}}", "{{{owner}}}"]
}

resource "elasticstack_elasticsearch_ingest_pipeline" "pipeline_a" {
  name = "pipeline_a"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_append.append_tags.json
  ]
}

data "elasticstack_elasticsearch_ingest_processor_fingerprint" "fingerprint" {
  fields = ["owner"]
}

// use the above defined pipeline in our configuration
data "elasticstack_elasticsearch_ingest_processor_pipeline" "pipeline" {
  name = elasticstack_elasticsearch_ingest_pipeline.pipeline_a.name
}

resource "elasticstack_elasticsearch_ingest_pipeline" "pipeline_b" {
  name = "pipeline_b"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_pipeline.pipeline.json,
    data.elasticstack_elasticsearch_ingest_processor_fingerprint.fingerprint.json
  ]
}
