provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_reroute" "test" {
  destination = "logs-app-default"
  dataset     = "application"
  namespace   = "production"
  description = "Route application logs"
  if          = "ctx.service?.name != null"
  tag         = "reroute-app-logs"
}
