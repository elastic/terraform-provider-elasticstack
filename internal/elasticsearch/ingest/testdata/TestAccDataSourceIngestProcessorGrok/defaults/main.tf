provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_grok" "test" {
  field    = "event.original"
  patterns = ["%%{WORD:event.action}"]
}
