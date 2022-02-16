provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_fail" "fail" {
  if      = "ctx.tags.contains('production') != true"
  message = "The production tag is not present, found tags: {{{tags}}}"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "fail-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_fail.fail.json
  ]
}
