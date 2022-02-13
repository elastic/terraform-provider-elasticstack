provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_network_direction" "network_direction" {
  internal_networks = ["private"]
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "network-direction-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_network_direction.network_direction.json
  ]
}
