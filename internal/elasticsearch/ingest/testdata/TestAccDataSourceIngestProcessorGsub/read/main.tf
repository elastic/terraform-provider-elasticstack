provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_gsub" "test" {
  field       = "field1"
  pattern     = "\\."
  replacement = "-"
}
