provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_reroute" "test" {
  namespace = "staging"
  if        = "ctx.event?.dataset == 'staging'"
}
