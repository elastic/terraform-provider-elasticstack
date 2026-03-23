provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = "my-index-fail"
  deletion_protection = false
}

resource "elasticstack_kibana_slo" "test_slo" {
  name        = "failwhale"
  slo_id      = "id-failwhale"
  description = "fully sick SLO"

  histogram_custom_indicator {
    index = "my-index-fail"

    good {
      field       = "test"
      aggregation = "supdawg"
      filter      = "latency < 300"
      from        = 0
      to          = 10
    }

    total {
      field       = "test"
      aggregation = "supdawg"
    }

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
