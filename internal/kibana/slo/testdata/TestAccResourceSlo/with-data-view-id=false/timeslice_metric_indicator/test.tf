variable "name" {
  type = string
}

variable "data_view_id" {
  type    = string
  default = null
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

  timeslice_metric_indicator {
    index           = "my-index-${var.name}"
    data_view_id    = var.data_view_id
    timestamp_field = "@timestamp"
    metric {
      metrics {
        name        = "A"
        aggregation = "sum"
        field       = "latency"
      }
      equation   = "A"
      comparator = "GT"
      threshold  = 100
    }
  }

  tags = ["tag-1", "another_tag"]

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

  space_id = "default"

  settings = {
    sync_delay = "5m"
    frequency  = "5m"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
