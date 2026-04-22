provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_kv" "test" {
  field       = "message"
  field_split = " "
  value_split = "="

  exclude_keys = ["tags"]
  prefix       = "setting_"
}
