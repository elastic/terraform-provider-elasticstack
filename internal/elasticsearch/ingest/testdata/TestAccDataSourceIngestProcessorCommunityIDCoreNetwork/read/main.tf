provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_community_id" "test" {
  source_ip        = "source.address"
  source_port      = 12345
  destination_ip   = "destination.address"
  destination_port = 443
  target_field     = "network.community_id"
  seed             = 123
  ignore_missing   = true
  ignore_failure   = true
}
