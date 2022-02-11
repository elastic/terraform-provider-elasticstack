provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_drop" "drop" {
  if = "ctx.network_name == 'Guest'"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "drop-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_drop.drop.json
  ]
}
