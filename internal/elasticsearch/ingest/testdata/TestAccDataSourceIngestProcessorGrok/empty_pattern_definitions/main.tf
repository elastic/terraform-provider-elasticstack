provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_grok" "test" {
  field               = "message"
  patterns            = ["%%{WORD:status}"]
  pattern_definitions = {}
}
