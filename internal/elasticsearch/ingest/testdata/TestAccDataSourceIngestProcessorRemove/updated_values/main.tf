provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_remove" "test" {
  field = [
    "host.name",
    "user.name",
  ]
  description = "Remove sensitive host and user fields"
  if          = "ctx.host?.name != null"
  tag         = "remove-sensitive-fields"
}
