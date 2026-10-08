provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_settings" "test" {
  index                    = "test-index"
  routing_rebalance_enable = "invalid"
}
