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
  description = "invalid calendar duration"

  kql_custom_indicator {
    index           = "my-index-${var.name}"
    good            = "true"
    total           = "*"
    timestamp_field = "custom_timestamp"
  }

  time_window {
    duration = "30d"
    type     = "calendarAligned"
  }

  budgeting_method = "timeslices"

  objective {
    target           = 0.95
    timeslice_target = 0.95
    timeslice_window = "5m"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
