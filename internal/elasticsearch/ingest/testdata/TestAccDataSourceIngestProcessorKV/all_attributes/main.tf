provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_kv" "test" {
  description    = "Parse selected labels"
  field          = "log.original"
  field_split    = "&"
  value_split    = ":"
  target_field   = "labels"
  include_keys   = ["env", "region"]
  exclude_keys   = ["debug"]
  ignore_missing = true
  prefix         = "kv_"
  trim_key       = "_"
  trim_value     = "|"
  strip_brackets = true
  if             = "ctx.log?.original != null"
  ignore_failure = true
  tag            = "kv-all-attributes"

  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "kv failed"
      }
    })
  ]
}
