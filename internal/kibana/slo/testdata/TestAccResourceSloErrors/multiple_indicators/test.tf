provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = "my-index-fail"
  deletion_protection = false
}

resource "elasticstack_kibana_slo" "test_slo" {
  name        = "fail"
  description = "multiple indicator fail"

  histogram_custom_indicator {
    index = "my-index-fail"

    good {
      field       = "test"
      aggregation = "value_count"
      filter      = "latency < 300"
    }

    total {
      field       = "test"
      aggregation = "value_count"
    }

    filter          = "labels.groupId: group-0"
    timestamp_field = "custom_timestamp"
  }

  kql_custom_indicator {
    index           = "my-index-fail"
    good            = "latency < 300"
    total           = "*"
    filter          = "labels.groupId: group-0"
    timestamp_field = "custom_timestamp"
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  budgeting_method = "timeslices"

  objective {
    target           = 0.999
    timeslice_target = 0.95
    timeslice_window = "5m"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
