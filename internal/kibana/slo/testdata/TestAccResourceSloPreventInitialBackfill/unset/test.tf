variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = "my-index-${var.name}"
  deletion_protection = false
}

resource "elasticstack_kibana_slo" "test_slo" {
  name        = var.name
  slo_id      = "id-${var.name}"
  description = "fully sick SLO"

  metric_custom_indicator {
    index = "my-index-${var.name}"
    good {
      metrics {
        name        = "A"
        aggregation = "sum"
        field       = "processor.processed"
      }
      metrics {
        name        = "B"
        aggregation = "sum"
        field       = "processor.processed"
      }
      equation = "A + B"
    }
    total {
      metrics {
        name        = "A"
        aggregation = "sum"
        field       = "processor.accepted"
      }
      metrics {
        name        = "B"
        aggregation = "sum"
        field       = "processor.accepted"
      }
      equation = "A + B"
    }
  }

  group_by = ["some.field", "some.other.field"]

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

  settings {
    sync_delay = "5m"
    frequency  = "5m"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
