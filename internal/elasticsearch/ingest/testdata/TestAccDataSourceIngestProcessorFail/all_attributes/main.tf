provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_fail" "test" {
  message        = "Document is missing a required deployment identifier"
  description    = "Fail when deployment metadata is missing"
  if             = "ctx.deployment_id == null"
  ignore_failure = true
  tag            = "fail-missing-deployment-id"
}
