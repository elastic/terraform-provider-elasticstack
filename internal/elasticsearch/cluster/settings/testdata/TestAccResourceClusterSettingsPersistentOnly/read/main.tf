provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_cluster_settings" "test_persistent" {
  persistent {
    setting {
      name  = "indices.lifecycle.poll_interval"
      value = "10m"
    }
  }
}
