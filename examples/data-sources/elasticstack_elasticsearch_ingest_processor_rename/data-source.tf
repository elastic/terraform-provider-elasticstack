provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_rename" "rename" {
  field        = "provider"
  target_field = "cloud.provider"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "rename-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_rename.rename.json
  ]
}
