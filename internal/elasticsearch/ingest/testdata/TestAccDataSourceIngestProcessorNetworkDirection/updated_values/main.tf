provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_network_direction" "test" {
  source_ip               = "src.address"
  destination_ip          = "dst.address"
  target_field            = "net.direction"
  internal_networks_field = "network.trusted_ranges"
}
