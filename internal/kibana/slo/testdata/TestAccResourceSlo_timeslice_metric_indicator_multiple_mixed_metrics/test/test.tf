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
  description = "multiple mixed metrics"

  timeslice_metric_indicator {
    index           = "my-index"
    timestamp_field = "@timestamp"

    metric {
      metrics {
        name        = "A"
        aggregation = "avg"
        field       = "bops"
      }
      metrics {
        name        = "B"
        aggregation = "percentile"
        field       = "latency"
        percentile  = 99
      }
      metrics {
        name        = "C"
        aggregation = "doc_count"
      }

      equation   = "A + B + C"
      comparator = "GT"
      threshold  = 100
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
