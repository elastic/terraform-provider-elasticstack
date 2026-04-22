provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_drop" "first" {
  description = "Equivalent drop processor"
  if          = "ctx.network_name == 'Guest'"
  tag         = "drop-tag"
}

data "elasticstack_elasticsearch_ingest_processor_drop" "second" {
  description = "Equivalent drop processor"
  if          = "ctx.network_name == 'Guest'"
  tag         = "drop-tag"
}
