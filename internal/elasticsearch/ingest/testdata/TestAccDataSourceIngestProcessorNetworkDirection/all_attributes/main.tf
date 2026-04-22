provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_network_direction" "test" {
  source_ip         = "source.ip"
  destination_ip    = "destination.ip"
  target_field      = "network.direction"
  internal_networks = ["private", "loopback"]
  description       = "Infer direction for private and loopback traffic"
  if                = "ctx.source?.ip != null && ctx.destination?.ip != null"
  ignore_failure    = true
  tag               = "network-direction"
}
