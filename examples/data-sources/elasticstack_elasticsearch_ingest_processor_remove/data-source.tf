provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_remove" "remove" {
  field = ["user_agent", "url"]
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "remove-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_remove.remove.json
  ]
}
