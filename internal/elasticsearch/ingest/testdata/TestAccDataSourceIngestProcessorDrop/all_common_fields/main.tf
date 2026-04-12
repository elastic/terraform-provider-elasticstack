provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_drop" "test" {
  description    = "Drop guest traffic"
  if             = "ctx.network_name == 'Guest'"
  ignore_failure = true
  tag            = "drop-guest-tag"
}
