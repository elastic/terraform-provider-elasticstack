provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_append" "tags" {
  field = "tags"
  value = ["failed"]
}

resource "elasticstack_elasticsearch_ingest_pipeline" "pipeline_with_failure_handler" {
  name = "pipeline_with_failure_handler"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_append.tags.json
  ]
}

data "elasticstack_elasticsearch_ingest_processor_pipeline" "test" {
  name = elasticstack_elasticsearch_ingest_pipeline.pipeline_with_failure_handler.name

  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "pipeline processor failed"
      }
    })
  ]
}
