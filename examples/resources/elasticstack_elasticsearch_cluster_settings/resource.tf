resource "elasticstack_elasticsearch_cluster_settings" "my_cluster_settings" {
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
      name  = "indices.breaker.accounting.limit"
      value = "100%"
    }
    setting {
      name       = "xpack.security.audit.logfile.events.include"
      value_list = ["ACCESS_DENIED", "ACCESS_GRANTED"]
    }
  }

  transient {
    setting {
      name  = "indices.breaker.accounting.limit"
      value = "99%"
    }
  }
}
