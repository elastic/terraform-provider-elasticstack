provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_script" "test" {
  description = "Run stored script to derive tags"
  lang        = "painless"
  script_id   = "stored-script-derive-tags"
}
