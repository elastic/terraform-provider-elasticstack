provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_cluster_settings" "test" {
  persistent {
    setting {
      name  = "indices.lifecycle.poll_interval"
      value = "10m"
    }
    setting {
      name  = "indices.recovery.max_bytes_per_sec"
      value = "50mb"
    }
    setting {
      name  = "indices.breaker.total.limit"
      value = "65%"
    }
  }

  transient {
    setting {
      name  = "indices.breaker.total.limit"
      value = "70%"
    }
    setting {
      name       = "xpack.security.audit.logfile.events.include"
      value_list = ["ACCESS_DENIED", "AUTHENTICATION_SUCCESS"]
    }
  }
}
