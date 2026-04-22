provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_network_direction" "test" {
  internal_networks = ["private"]
}
