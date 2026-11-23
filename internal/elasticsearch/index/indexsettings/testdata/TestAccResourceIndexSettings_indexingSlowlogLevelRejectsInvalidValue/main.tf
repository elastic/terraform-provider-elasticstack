provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_settings" "test" {
  index                  = "test-index"
  indexing_slowlog_level = "invalid"
}
