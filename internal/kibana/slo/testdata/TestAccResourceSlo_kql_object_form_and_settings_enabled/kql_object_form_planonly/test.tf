# Plan-only step: object-form KQL (filter_kql, good_kql) and settings sync_field parse; no apply.
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
  description = "plan only kql object form"
  enabled     = true

  kql_custom_indicator {
    index = "my-index-${var.name}"
    filter_kql = {
      kql_query = "service.name: test"
    }
    good_kql = {
      kql_query = "latency < 300"
    }
    total           = "*"
    timestamp_field = "custom_timestamp"
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  budgeting_method = "timeslices"

  objective {
    target           = 0.95
    timeslice_target = 0.95
    timeslice_window = "5m"
  }

  space_id = "default"

  settings {
    sync_delay = "5m"
    frequency  = "5m"
    sync_field = "@timestamp"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
