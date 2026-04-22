provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_dot_expander" "test" {
  field          = "foo.bar"
  path           = "nested"
  override       = true
  description    = "Expand dot fields"
  if             = "ctx.foo != null"
  ignore_failure = true
  tag            = "dot-expander-tag"
}
