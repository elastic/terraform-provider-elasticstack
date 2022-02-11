provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_community_id" "community" {}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "community-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_community_id.community.json
  ]
}
