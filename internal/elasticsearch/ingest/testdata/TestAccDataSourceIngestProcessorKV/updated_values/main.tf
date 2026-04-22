provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_kv" "test" {
  field        = "labels"
  field_split  = ";"
  value_split  = "=>"
  target_field = "parsed_labels"
  include_keys = ["service", "zone"]
  exclude_keys = ["debug", "temp"]
  prefix       = "meta_"
  trim_key     = "-"
  trim_value   = "~"
}
