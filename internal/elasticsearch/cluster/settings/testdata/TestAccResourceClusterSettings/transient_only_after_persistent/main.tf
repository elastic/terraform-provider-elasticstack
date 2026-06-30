provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_cluster_settings" "test" {
  transient {
    setting {
      name  = "indices.breaker.total.limit"
      value = "55%"
    }
  }
}
