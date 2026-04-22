provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_append" "test" {
  description      = "Append tags to the doc"
  field            = "tags"
  value            = ["production", "{{{app}}}", "{{{owner}}}"]
  allow_duplicates = true
}
