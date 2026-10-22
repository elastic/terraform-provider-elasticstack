variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = "my-index"
  deletion_protection = false
}

resource "elasticstack_kibana_slo" "test_slo" {
  name        = var.name
  description = "std_deviation timeslice metric"
  timeslice_metric_indicator {
    index           = "my-index"
    timestamp_field = "@timestamp"
    metric {
      metrics {
        name        = "A"
        aggregation = "std_deviation"
        field       = "latency"
      }
      equation   = "A"
      comparator = "GT"
      threshold  = 10
    }
  }
  budgeting_method = "timeslices"
  objective {
    target           = 0.95
    timeslice_target = 0.95
    timeslice_window = "5m"
  }
  time_window {
    duration = "7d"
    type     = "rolling"
  }
  depends_on = [elasticstack_elasticsearch_index.my_index]
}
