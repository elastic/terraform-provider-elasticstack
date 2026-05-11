provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_cluster_settings" "test" {
  persistent {
    setting {
      name  = "indices.lifecycle.poll_interval"
      value = "15m"
    }
    setting {
      name  = "indices.recovery.max_bytes_per_sec"
      value = "40mb"
    }
    setting {
      name  = "indices.breaker.total.limit"
      value = "60%"
    }
    setting {
      name       = "xpack.security.audit.logfile.events.include"
      value_list = ["ACCESS_DENIED", "ACCESS_GRANTED"]
    }
  }
}
