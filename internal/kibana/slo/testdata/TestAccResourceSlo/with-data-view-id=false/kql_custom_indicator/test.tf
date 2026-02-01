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

  kql_custom_indicator {
    index           = "my-index-${var.name}"
    data_view_id    = var.data_view_id
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

  space_id = "default"

  settings = {
    sync_delay = "5m"
    frequency  = "5m"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
