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
  description = "doc_count timeslice metric"

  timeslice_metric_indicator {
    index           = "my-index"
    timestamp_field = "@timestamp"

    metric {
      metrics {
        name        = "C"
        aggregation = "doc_count"
      }
      equation   = "C"
      comparator = "GTE"
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
