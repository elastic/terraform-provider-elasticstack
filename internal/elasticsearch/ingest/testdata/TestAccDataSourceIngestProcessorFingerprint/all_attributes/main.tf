provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_fingerprint" "test" {
  fields         = ["user", "email", "ip"]
  target_field   = "doc_fingerprint"
  method         = "SHA-256"
  salt           = "my-secret-salt"
  ignore_missing = true
  description    = "Fingerprint for dedup"
  if             = "ctx.env == 'prod'"
  ignore_failure = true
  tag            = "fingerprint-docs"
}
