provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_append" "tags" {
  field = "tags"
  value = ["metadata", "{{{service.name}}}"]
}

resource "elasticstack_elasticsearch_ingest_pipeline" "pipeline_with_metadata" {
  name = "pipeline_with_metadata"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_append.tags.json
  ]
}

data "elasticstack_elasticsearch_ingest_processor_pipeline" "test" {
  name           = elasticstack_elasticsearch_ingest_pipeline.pipeline_with_metadata.name
  description    = "Route documents through the metadata pipeline"
  if             = "ctx.service?.name != null"
  ignore_failure = true
  tag            = "pipeline-metadata-tag"
}
