provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_geoip" "test" {
  field          = "ip"
  target_field   = "geoip"
  ignore_missing = true
  ignore_failure = true
  description    = "geoip lookup"
  if             = "ctx.ip != null"
  tag            = "geoip-tag"
  on_failure     = ["{\"set\":{\"field\":\"error.message\",\"value\":\"geoip failed\"}}"]
}
